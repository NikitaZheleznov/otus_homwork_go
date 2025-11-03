package hw03frequencyanalysis

import (
	"sort"
	"strings"
	"unicode"
)

type WordEntity struct {
	Word  string
	Count int
}

func Top10(text string) []string {
	wordMap := make(map[string]int)
	for _, word := range strings.Fields(text) {
		w := strings.ToLower(strings.TrimFunc(word, unicode.IsPunct))
		if w != "" {
			wordMap[w]++
		}
	}
	count := 10
	if count > len(wordMap) {
		count = len(wordMap)
	}
	words := make([]WordEntity, 0, count)
	for k, v := range wordMap {
		words = append(words, WordEntity{
			Word:  k,
			Count: v,
		})
	}
	sort.Slice(words, func(i, j int) bool {
		if words[i].Count == words[j].Count {
			return words[i].Word < words[j].Word
		}
		return words[i].Count > words[j].Count
	})
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, words[i].Word)
	}
	return result
}
