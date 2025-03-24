package fuzhu

import (
	"regexp"
	"sync"
)

// 正则表达式管理器
type RegexManager struct {
	regexps []*regexp.Regexp
	mu      sync.RWMutex
}

func NewRegexManager() *RegexManager {
	return &RegexManager{
		regexps: make([]*regexp.Regexp, 0, 100),
	}
}

func (rm *RegexManager) AddPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rm.mu.Lock()
	rm.regexps = append(rm.regexps, re)
	rm.mu.Unlock()
	return nil
}

func (rm *RegexManager) MatchAll(data []byte) []string {
	var matches []string
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	matchChan := make(chan string, len(rm.regexps))
	var wg sync.WaitGroup

	for _, re := range rm.regexps {
		wg.Add(1)
		go func(re *regexp.Regexp) {
			defer wg.Done()
			if re.Match(data) {
				matchChan <- re.String()
			}
		}(re)
	}

	go func() {
		wg.Wait()
		close(matchChan)
	}()

	for match := range matchChan {
		matches = append(matches, match)
	}

	return matches
}
