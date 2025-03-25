package fuzhu

import (
	"sync"

	"github.com/cloudflare/ahocorasick"
)

type StringManager struct {
	patterns []string
	matcher  *ahocorasick.Matcher
	mu       sync.RWMutex
}

func NewStringManager() *StringManager {
	return &StringManager{
		patterns: make([]string, 0, 100),
	}
}
func (rm *StringManager) AddStringPattern(pattern string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.patterns = append(rm.patterns, pattern)
	// 重新构建匹配器
	rm.matcher = ahocorasick.NewStringMatcher(rm.patterns)
	return nil
}
func (rm *StringManager) MatchAllStrings(data []byte) []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.matcher == nil {
		return nil
	}

	// 执行匹配
	matches := rm.matcher.Match(data)
	results := make([]string, 0, len(matches))

	// 将匹配结果转换为原始模式
	for _, idx := range matches {
		results = append(results, rm.patterns[idx])
	}

	return results
}
