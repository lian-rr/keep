package reverse

import (
	"strings"

	"github.com/kljensen/snowball"
)

const englishLang = "english"

type basicTokenizer struct {
	minLenght int
}

func (t basicTokenizer) tokenizeStr(str string, exl set) set {
	words := strings.Fields(str)
	return t.tokenize(words, exl)
}

func (t basicTokenizer) tokenize(words []string, exl set) set {
	tokens := make(set)
	if exl == nil {
		exl = make(set)
	}

	for _, word := range words {
		if len(word) < t.minLenght {
			continue
		}

		if _, ok := exl[word]; !ok {
			tokens[word] = struct{}{}
		}
	}

	return tokens
}

func (t basicTokenizer) stemsFilter(tokens set) (set, error) {
	stems := make(set)
	for token := range tokens {
		stem, err := snowball.Stem(token, englishLang, false)
		if err != nil {
			return nil, err
		}

		stems[stem] = struct{}{}
	}

	return stems, nil
}
