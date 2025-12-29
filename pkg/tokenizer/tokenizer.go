package tokenizer

import (
	"regexp"
	"strings"
)

type TokenType int

const (
	RAWHTML TokenType = iota
	INCLUDE
)

type Token struct {
	Type    TokenType
	Content string
}

var includeRegex = regexp.MustCompile(`{{\s*include\s+(.+?)\s*}}`)

func Tokenize(content string) []Token {
	var tokens []Token

	matches := includeRegex.FindAllStringSubmatchIndex(content, -1)
	lastIndex := 0

	for _, match := range matches {
		if match[0] > lastIndex {
			tokens = append(tokens, Token{Type: RAWHTML, Content: content[lastIndex:match[0]]})
		}

		includePath := content[match[2]:match[3]]
		tokens = append(tokens, Token{Type: INCLUDE, Content: strings.TrimSpace(includePath)})

		lastIndex = match[1]
	}

	if lastIndex < len(content) {
		tokens = append(tokens, Token{Type: RAWHTML, Content: content[lastIndex:]})
	}

	return tokens
}
