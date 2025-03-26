package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gopr/fuzhu"
	"gopr/fuzhu/logger"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

var regexManager = fuzhu.NewRegexManager()

func main() {

	// 添加命令行参数支持
	upstreamProxyFlag := flag.String("p", "", "上游代理地址 (例如: http://proxy:port)")
	verboseFlag := flag.Bool("v", false, "详细信息")
	flag.Parse()

	go processResponseLogs()
	// +++初始化正则表达式管理器+++

	for _, pattern := range fuzhu.SecretPatterns {
		if err := regexManager.AddPattern(pattern); err != nil {
			logger.Errorf("添加正则表达式失败: %v", err)
		}
	}
	// ---初始化正则表达式管理器---

	proxyServer := goproxy.NewProxyHttpServer()
	proxyServer.Verbose = *verboseFlag

	// 根据命令行参数配置上游代理
	if *upstreamProxyFlag != "" {
		upstreamProxy, err := url.Parse(*upstreamProxyFlag)
		if err != nil {
			logger.Fatal("解析上游代理地址失败:", err)
		}

		proxyServer.Tr = &http.Transport{
			Proxy: http.ProxyURL(upstreamProxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// 对于 HTTP 代理，我们需要直接使用 net.Dialer
				dialer := &net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}

				if upstreamProxy.Scheme == "http" || upstreamProxy.Scheme == "https" {
					return dialer.DialContext(ctx, network, addr)
				}

				// 对于 SOCKS 代理
				proxyDialer, err := proxy.FromURL(upstreamProxy, dialer)
				if err != nil {
					return nil, err
				}
				return proxyDialer.Dial(network, addr)
			},
			// 添加以下优化配置
			MaxIdleConns:        1000,             // 最大空闲连接数
			MaxIdleConnsPerHost: 100,              // 每个主机的最大空闲连接数
			MaxConnsPerHost:     100,              // 每个主机的最大连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			DisableKeepAlives:   false,            // 启用 keep-alive
		}
		logger.Infof("已启用上游代理: %s", *upstreamProxyFlag)
	} else {
		// 即使不使用上游代理，也优化本地代理的传输设置
		proxyServer.Tr = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		}
	}

	// 加载自定义证书
	caCert, err := os.ReadFile("ca.crt")
	if err != nil {
		logger.Fatal("读取CA证书失败:", err)
	}
	caKey, err := os.ReadFile("ca.key")
	if err != nil {
		logger.Fatal("读取CA私钥失败:", err)
	}

	// 设置HTTPS支持
	ca, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		logger.Fatal("加载证书失败:", err)
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		logger.Fatal("解析证书失败:", err)
	}
	goproxy.GoproxyCa = ca
	proxyServer.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 监听所有请求
	proxyServer.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// if req.Header.Get("Upgrade") == "websocket" {
		// 	// 读取 WebSocket 请求体
		// 	body, err := io.ReadAll(req.Body)
		// 	if err == nil && len(body) > 0 {
		// 		logger.Printf("[WebSocket发送] %s %s\n内容: %s\n", req.Method, req.URL, string(body))
		// 		// 重新设置请求体，因为已经被读取
		// 		req.Body = io.NopCloser(bytes.NewBuffer(body))
		// 	}
		// }
		// logger.Printf("[请求] %s %s\n", req.Method, req.URL)
		return req, nil
	})

	// 监听所有响应
	proxyServer.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil || ctx == nil || ctx.Req == nil {
			return resp
		}
		domain := extractMainDomain(ctx.Req.URL.Host)
		if shouldSkipHost(domain) {
			// logger.Debugf("skip host: %s", domain)
			return resp
		}

		contentType := resp.Header.Get("Content-Type")
		if shouldSkipContentType(contentType) {
			return resp
		}
		if resp.StatusCode != 200 {
			return resp
		}
		// 读取响应体
		var body []byte
		if resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
			// 重新设置响应体
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		// 将数据发送到队列
		select {
		case responseQueue <- ResponseData{
			Method:     ctx.Req.Method,
			URL:        ctx.Req.URL.String(),
			StatusCode: resp.StatusCode,
			Body:       body,
		}:
		default:
			// 如果队列满了，记录一个警告但不阻塞
			logger.Warn("Response queue is full, skipping log")
		}
		// if false {
		// 	if resp != nil {
		// 		body, err := io.ReadAll(resp.Body)
		// 		if err == nil {
		// 			// 进行正则匹配
		// 			if matches := regexManager.MatchAll(body); len(matches) > 0 {
		// 				logger.Printf("匹配到的正则: %v", matches)
		// 			}
		// 			// 重新设置响应体
		// 			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		// 		}
		// 	}
		// }
		// if false {
		// 	if resp != nil {
		// 		// 读取响应体
		// 		body, err := io.ReadAll(resp.Body)
		// 		if err == nil {
		// 			// 检查是否为图片
		// 			if fuzhu.IsImageResponse(resp) {
		// 				if err := fuzhu.SaveImage(body, resp.Request.URL.String()); err == nil {
		// 					logger.Printf("已保存图片: %s\n", resp.Request.URL)
		// 				}
		// 			}
		// 			// 重新设置响应体
		// 			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		// 		}
		// 	}
		// }
		return resp
	})

	logger.Print("启动代理服务器在 :8889...")
	logger.Fatal(http.ListenAndServe(":8889", proxyServer))
}

type ResponseData struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

var responseQueue = make(chan ResponseData, 100000)

func processResponseLogs() {
	for data := range responseQueue {
		matches := regexManager.MatchAll(data.Body)
		if len(matches) > 0 {
			for i := 0; i < len(matches); i++ {
				m := matches[i]
				if strings.Contains(m.GroupValues[0], `"same-origin"`) {
					continue
				}
				logger.Infof("[res] %s %s -> %v", data.Method, data.URL, m.GroupValues[0])
			}
		}
		// logger.Printf("[res] %s %s -> [%d] %d", data.Method, data.URL, data.StatusCode, bodylength)
	}
}
func shouldSkipContentType(contentType string) bool {
	skipTypes := []string{
		"image/",
		"font/",
		"text/css",
		"video/",
		"audio/",
		"application/font",
		"application/x-font",
	}

	for _, skipType := range skipTypes {
		if strings.HasPrefix(contentType, skipType) {
			return true
		}
	}
	return false
}
func shouldSkipHost(host string) bool {
	skipHosts := map[string]bool{
		"google.com":     true,
		"gstatic.com":    true,
		"googleapis.com": true,
		"github.com":     true,
		"cloudflare.com": true,
		"gravatar.com":   true,
		"youtube.com":    true,
		"ytimg.com":      true,
		"facebook.com":   true,
		"fbcdn.net":      true,
		"twitter.com":    true,
		"twimg.com":      true,
		"microsoft.com":  true,
		"msn.com":        true,
		"live.com":       true,
		"akamai.net":     true,
		"jsdelivr.net":   true,
		"unpkg.com":      true,
		"baidu.com":      true,
		"csdn.net":       true,
		"cnblogs.com":    true,
	}

	mainDomain := extractMainDomain(host)
	return skipHosts[mainDomain]
}
func extractMainDomain(host string) string {
	// 移除端口号
	if i := strings.Index(host, ":"); i != -1 {
		host = host[:i]
	}
	parts := strings.Split(host, ".")
	length := len(parts)

	// 处理无效域名
	if length < 2 {
		return host
	}

	// 获取主域名部分
	return parts[length-2] + "." + parts[length-1]
}
