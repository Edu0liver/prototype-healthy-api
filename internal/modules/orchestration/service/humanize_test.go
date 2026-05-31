package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripMarkdown(t *testing.T) {
	require.Equal(t, "click here (https://x.com)", stripMarkdown("[click here](https://x.com)"))
	require.Equal(t, "bold", stripMarkdown("**bold**"))
	require.Equal(t, "Title", stripMarkdown("## Title"))
	require.Equal(t, "code", stripMarkdown("`code`"))
}

func TestHumanize_TrimsAndDropsEmpty(t *testing.T) {
	out := humanize("  hello  ")
	require.Equal(t, []string{"hello"}, out)
	require.Empty(t, humanize("   "))
}

func TestHumanize_CapsAtMaxMessages(t *testing.T) {
	// 10 paragraphs must collapse into at most maxOutboundMessages buckets.
	text := strings.Repeat("para\n\n", 10)
	out := humanize(text)
	require.LessOrEqual(t, len(out), maxOutboundMessages)
	require.NotEmpty(t, out)
}

func TestSplitNonEmpty(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, splitNonEmpty("a\n\n\nb\n", "\n"))
	require.Empty(t, splitNonEmpty("   ", "\n"))
}
