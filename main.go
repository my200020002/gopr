package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"gopr/fuzhu"
	"gopr/fuzhu/logger"

	"github.com/elazarl/goproxy"
)

func main() {

	// 添加命令行参数支持
	upstreamProxyFlag := flag.String("p", "", "上游代理地址 (例如: http://proxy:port)")
	verboseFlag := flag.Bool("v", false, "详细信息")
	flag.Parse()
	// +++初始化正则表达式管理器+++
	regexManager := fuzhu.NewRegexManager()
	patterns := []string{
		`pattern1`,
		`pattern2`,
		// ... 更多模式
	}
	for _, pattern := range patterns {
		if err := regexManager.AddPattern(pattern); err != nil {
			logger.Errorf("添加正则表达式失败: %v", err)
		}
	}
	// ---初始化正则表达式管理器---

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verboseFlag

	// 根据命令行参数配置上游代理
	if *upstreamProxyFlag != "" {
		upstreamProxy, err := url.Parse(*upstreamProxyFlag)
		if err != nil {
			logger.Fatal("解析上游代理地址失败:", err)
		}

		proxy.Tr = &http.Transport{
			Proxy: http.ProxyURL(upstreamProxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			// 添加以下优化配置
			MaxIdleConns:        1000,             // 最大空闲连接数
			MaxIdleConnsPerHost: 100,              // 每个主机的最大空闲连接数
			MaxConnsPerHost:     100,              // 每个主机的最大连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			DisableKeepAlives:   false,            // 启用 keep-alive
		}
		logger.Infof("已启用上游代理: %s\n", *upstreamProxyFlag)
	} else {
		// 即使不使用上游代理，也优化本地代理的传输设置
		proxy.Tr = &http.Transport{
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
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 监听所有请求
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if req.Header.Get("Upgrade") == "websocket" {
			// 读取 WebSocket 请求体
			body, err := io.ReadAll(req.Body)
			if err == nil && len(body) > 0 {
				logger.Printf("[WebSocket发送] %s %s\n内容: %s\n", req.Method, req.URL, string(body))
				// 重新设置请求体，因为已经被读取
				req.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
		// logger.Printf("[请求] %s %s\n", req.Method, req.URL)
		return req, nil
	})

	// 监听所有响应
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil && resp.Header.Get("Upgrade") == "websocket" {
			// 读取 WebSocket 响应体
			body, err := io.ReadAll(resp.Body)
			if err == nil && len(body) > 0 {
				logger.Printf("[WebSocket接收] %s %s -> %d\n内容: %s\n",
					ctx.Req.Method, ctx.Req.URL, resp.StatusCode, string(body))
				// 重新设置响应体
				resp.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
		if false {
			if resp != nil {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					// 进行正则匹配
					if matches := regexManager.MatchAll(body); len(matches) > 0 {
						logger.Printf("匹配到的正则: %v", matches)
					}
					// 重新设置响应体
					resp.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}
		}
		if false {
			if resp != nil {
				// 读取响应体
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					// 检查是否为图片
					if fuzhu.IsImageResponse(resp) {
						if err := fuzhu.SaveImage(body, resp.Request.URL.String()); err == nil {
							logger.Printf("已保存图片: %s\n", resp.Request.URL)
						}
					}
					// 重新设置响应体
					resp.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}
		}

		logger.Printf("[响应] %s %s -> %d\n", ctx.Req.Method, ctx.Req.URL, resp.StatusCode)
		return resp
	})

	logger.Print("启动代理服务器在 :8889...")
	logger.Fatal(http.ListenAndServe(":8889", proxy))
}
