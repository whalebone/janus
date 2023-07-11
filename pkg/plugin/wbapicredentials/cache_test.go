package wbapicredentials

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContainsTrue(t *testing.T) {
	c := NewCache(5*time.Minute, 5*time.Minute)
	c.Put("key")
	require.True(t, c.Contains("key"))
}

func TestContainsFalse(t *testing.T) {
	c := NewCache(5*time.Minute, 5*time.Minute)
	c.Put("key")
	require.False(t, c.Contains("key2"))
}

func TestExpiration(t *testing.T) {
	c := NewCache(1*time.Second, 5*time.Minute)
	c.Put("key")
	found := c.Contains("key")
	require.True(t, found)

	// record should be still cached
	time.Sleep(900 * time.Millisecond)
	found = c.Contains("key")
	require.True(t, found)

	// record should be expired regardless it was accessed several millis ago
	time.Sleep(100 * time.Millisecond)
	found = c.Contains("key")
	require.False(t, found)
}

func TestFlush(t *testing.T) {
	c := NewCache(5*time.Minute, 5*time.Minute)
	c.Put("key")
	found := c.Contains("key")
	require.True(t, found)

	c.Flush()

	found = c.Contains("key")
	require.False(t, found)
}
