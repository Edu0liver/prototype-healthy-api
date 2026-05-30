package service

import (
	"regexp"
	"strings"
)

const maxOutboundMessages = 4

var (
	mdLinkRe   = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)]+)\)`)
	mdBoldRe   = regexp.MustCompile(`(\*\*|__)(.*?)(\*\*|__)`)
	mdItalicRe = regexp.MustCompile(`(?m)(^|\s)[*_]([^*_\n]+)[*_]`)
	mdHeadRe   = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	mdCodeRe   = regexp.MustCompile("`{1,3}")
)

// humanize converts an LLM answer into up to N chat-friendly messages: strips
// markdown unsuited to chat, converts links to "text (url)", and splits long
// answers (PROMPT 7).
func humanize(text string) []string {
	text = stripMarkdown(text)
	parts := splitMessages(text, maxOutboundMessages)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stripMarkdown(s string) string {
	s = mdLinkRe.ReplaceAllString(s, "$1 ($2)")
	s = mdBoldRe.ReplaceAllString(s, "$2")
	s = mdItalicRe.ReplaceAllString(s, "$1$2")
	s = mdHeadRe.ReplaceAllString(s, "")
	s = mdCodeRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// splitMessages divides text into at most max chunks along paragraph/sentence
// boundaries, keeping each chunk reasonably small.
func splitMessages(text string, max int) []string {
	paras := splitNonEmpty(text, "\n\n")
	if len(paras) <= 1 {
		paras = splitNonEmpty(text, "\n")
	}
	if len(paras) <= max {
		return paras
	}
	// Too many paragraphs: greedily merge into `max` buckets.
	buckets := make([]string, max)
	for i, p := range paras {
		idx := i * max / len(paras)
		if idx >= max {
			idx = max - 1
		}
		if buckets[idx] != "" {
			buckets[idx] += "\n\n"
		}
		buckets[idx] += p
	}
	return buckets
}

func splitNonEmpty(text, sep string) []string {
	raw := strings.Split(text, sep)
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		if s := strings.TrimSpace(r); s != "" {
			out = append(out, s)
		}
	}
	return out
}
