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
	PROP
)

type Token struct {
	Type    TokenType
	Content string // For directives, this is the argument string
    Start   int
    End     int
}

// Matches {{ directive args... }}
// Group 1: directive
// Group 2: args (optional)
var directiveRegex = regexp.MustCompile(`{{\s*(\w+)\s*(.*?)\s*}}`)

func Tokenize(content string) []Token {
	var tokens []Token

	matches := directiveRegex.FindAllStringSubmatchIndex(content, -1)
	lastIndex := 0

	for _, match := range matches {
		// Append RAWHTML before the tag
		if match[0] > lastIndex {
			tokens = append(tokens, Token{
                Type: RAWHTML, 
                Content: content[lastIndex:match[0]],
                Start: lastIndex,
                End: match[0],
            })
		}

		// Extract directive and arguments
		directive := content[match[2]:match[3]]
		args := ""
		if match[4] != -1 {
			args = content[match[4]:match[5]]
		}

        token := Token{Content: args, Start: match[0], End: match[1]}

		switch directive {
		case "include":
            token.Type = INCLUDE
		case "layout":
            token.Type = LAYOUT
		case "endlayout":
            token.Type = ENDLAYOUT
		case "block":
            token.Type = BLOCK
		case "endblock":
            token.Type = ENDBLOCK
		case "slot":
            token.Type = SLOT
		case "prop":
            token.Type = PROP
		default:
			// Treat unknown directives as RAWHTML
            token.Type = RAWHTML
            token.Content = content[match[0]:match[1]]
		}
        tokens = append(tokens, token)

		lastIndex = match[1]
	}

	if lastIndex < len(content) {
		tokens = append(tokens, Token{
            Type: RAWHTML, 
            Content: content[lastIndex:],
            Start: lastIndex,
            End: len(content),
        })
	}

	return tokens
}
