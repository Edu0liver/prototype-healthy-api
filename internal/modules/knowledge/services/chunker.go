package services

import (
	"regexp"
	"strings"
)

// approxCharsPerToken is a rough heuristic to map a token budget to characters.
const approxCharsPerToken = 4

// chunkText splits text into overlapping chunks sized by an approximate token
// budget (chunkSize/overlap are in tokens; converted to characters).
func chunkText(text string, chunkSizeTokens, overlapTokens int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	size := chunkSizeTokens * approxCharsPerToken
	overlap := overlapTokens * approxCharsPerToken
	if size <= 0 {
		size = 3200
	}
	if overlap < 0 || overlap >= size {
		overlap = size / 8
	}

	runes := []rune(text)
	var chunks []string
	step := size - overlap
	if step <= 0 {
		step = size
	}
	for start := 0; start < len(runes); start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}

var (
	htmlTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	whitespaceRe = regexp.MustCompile(`[ \t]+`)
	blanklineRe  = regexp.MustCompile(`\n{3,}`)
)

// estimateTokens approximates token count from character length.
func estimateTokens(text string) int { return len([]rune(text)) / approxCharsPerToken }
