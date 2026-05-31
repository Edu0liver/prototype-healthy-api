package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkText_Empty(t *testing.T) {
	require.Nil(t, chunkText("   ", 800, 100))
}

func TestChunkText_SingleChunkWhenShort(t *testing.T) {
	chunks := chunkText("hello world", 800, 100)
	require.Len(t, chunks, 1)
	require.Equal(t, "hello world", chunks[0])
}

func TestChunkText_SplitsWithOverlap(t *testing.T) {
	// 10 tokens => 40 chars per chunk, 2 tokens => 8 chars overlap, step = 32 chars.
	text := strings.Repeat("a", 100)
	chunks := chunkText(text, 10, 2)
	require.Greater(t, len(chunks), 1, "long text must split into multiple chunks")
	for _, c := range chunks {
		require.LessOrEqual(t, len([]rune(c)), 40)
	}
}

func TestEstimateTokens(t *testing.T) {
	require.Equal(t, 3, estimateTokens("abcdefghijkl")) // 12 chars / 4
}
