package tokenizer

import (
	"regexp"
)

type TokenType int

const (
	RAWHTML TokenType = iota
	INCLUDE
	LAYOUT
	ENDLAYOUT
	BLOCK
	ENDBLOCK
	SLOT
)

type Token struct {
	Type    TokenType
	Content string // For directives, this is the argument (e.g. path or name)
}

// Matches {{ directive "argument" }} or {{ directive }}
// Group 1: directive
// Group 2: argument (optional, with quotes)
var directiveRegex = regexp.MustCompile(`{{\s*(\w+)(?:\s+"(.+?)")?\s*}}`)

func Tokenize(content string) []Token {
	var tokens []Token

	matches := directiveRegex.FindAllStringSubmatchIndex(content, -1)
	lastIndex := 0

	for _, match := range matches {
		// Append RAWHTML before the tag
		if match[0] > lastIndex {
			tokens = append(tokens, Token{Type: RAWHTML, Content: content[lastIndex:match[0]]})
		}

		// Extract directive and argument
		directive := content[match[2]:match[3]]
		arg := ""
		if match[4] != -1 {
			arg = content[match[4]:match[5]]
		}

		switch directive {
		case "include":
			tokens = append(tokens, Token{Type: INCLUDE, Content: arg})
		case "layout":
			tokens = append(tokens, Token{Type: LAYOUT, Content: arg})
		case "endlayout":
			tokens = append(tokens, Token{Type: ENDLAYOUT})
		case "block":
			tokens = append(tokens, Token{Type: BLOCK, Content: arg})
		case "endblock":
			tokens = append(tokens, Token{Type: ENDBLOCK})
		case "slot":
			tokens = append(tokens, Token{Type: SLOT, Content: arg})
		default:
			// Treat unknown directives as RAWHTML (or handle error? for now keep as text)
			tokens = append(tokens, Token{Type: RAWHTML, Content: content[match[0]:match[1]]})
		}

		lastIndex = match[1]
	}

	if lastIndex < len(content) {
		tokens = append(tokens, Token{Type: RAWHTML, Content: content[lastIndex:]})
	}

	return tokens
}
