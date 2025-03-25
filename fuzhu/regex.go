package fuzhu

import (
	"fmt"
	"regexp"
	"sync"
)

// 正则表达式管理器
type RegexManager struct {
	regexps []*regexp.Regexp
	mu      sync.RWMutex
}
type Match struct {
	Pattern     string            // 匹配的正则表达式
	Value       string            // 完整匹配内容
	Groups      map[string]string // 命名分组结果
	GroupValues []string          // 所有分组结果（包括未命名的）
	Index       int               // 匹配位置
	Length      int               // 匹配长度
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

func (rm *RegexManager) MatchAll(data []byte) []Match {
	var matches []Match
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	matchChan := make(chan []Match, len(rm.regexps))
	var wg sync.WaitGroup

	for _, re := range rm.regexps {
		wg.Add(1)
		go func(re *regexp.Regexp) {
			defer wg.Done()
			var results []Match

			groupNames := re.SubexpNames()
			allMatches := re.FindAllSubmatch(data, -1)
			allIndexes := re.FindAllSubmatchIndex(data, -1)

			// 使用map去重
			seen := make(map[string]bool)

			for i, submatch := range allMatches {
				// 跳过空匹配
				if len(submatch) == 0 || len(submatch[0]) == 0 {
					continue
				}

				// 检查是否重复
				matchKey := string(submatch[0])
				if seen[matchKey] {
					continue
				}
				seen[matchKey] = true

				match := Match{
					Pattern:     re.String(),
					Value:       matchKey,
					Groups:      make(map[string]string),
					GroupValues: make([]string, 0, len(submatch)),
					Index:       allIndexes[i][0],
					Length:      allIndexes[i][1] - allIndexes[i][0],
				}

				// 处理分组，跳过空分组
				for j, group := range submatch {
					if len(group) > 0 {
						match.GroupValues = append(match.GroupValues, string(group))
						if j > 0 && j < len(groupNames) && groupNames[j] != "" {
							match.Groups[groupNames[j]] = string(group)
						}
					}
				}

				results = append(results, match)
			}

			if len(results) > 0 {
				matchChan <- results
			}
		}(re)
	}

	go func() {
		wg.Wait()
		close(matchChan)
	}()

	for results := range matchChan {
		matches = append(matches, results...)
	}

	return matches
}
func (rm *RegexManager) MatchAllString(data string) []Match {
	return rm.MatchAll([]byte(data))
}
func PrintMatches(matches []Match) {
	for i, m := range matches {
		fmt.Printf("\n=== 匹配结果 #%d ===\n", i+1)
		fmt.Printf("Pattern: %s\n", m.Pattern)
		fmt.Printf("Value: %s\n", m.Value)
		fmt.Printf("Index: %d\n", m.Index)
		fmt.Printf("Length: %d\n", m.Length)

		fmt.Println("Groups:")
		for name, value := range m.Groups {
			fmt.Printf("  %s: %s\n", name, value)
		}

		fmt.Println("GroupValues:")
		for i, value := range m.GroupValues {
			fmt.Printf("  [%d]: %s\n", i, value)
		}
		fmt.Println("==================")
	}
}
