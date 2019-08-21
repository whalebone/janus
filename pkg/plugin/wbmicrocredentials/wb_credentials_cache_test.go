package wbmicrocredentials

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetFound(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", &CachedCredentials{LoginSuccess: true, ClientID: "1"})
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "1", cred.ClientID)
	require.True(t, cred.LoginSuccess)
}

func TestGetNotFound(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", &CachedCredentials{LoginSuccess: true, ClientID: "1"})
	cred, found := c.Get("key2")
	require.False(t, found)
	require.Nil(t, cred)
}

func TestReplace(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", &CachedCredentials{LoginSuccess: true, ClientID: "1"})
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "1", cred.ClientID)
	require.True(t, cred.LoginSuccess)
	c.Put("key", &CachedCredentials{LoginSuccess: false, ClientID: "3"})
	cred, found = c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "3", cred.ClientID)
	require.False(t, cred.LoginSuccess)
}

func TestExpiration(t *testing.T) {
	c := NewCredentialsCache(1*time.Second, 5*time.Minute)
	c.Put("key", &CachedCredentials{LoginSuccess: true, ClientID: "1"})
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "1", cred.ClientID)
	require.True(t, cred.LoginSuccess)

	// wait until the record expires
	time.Sleep(1 * time.Second)
	cred, found = c.Get("key")
	require.False(t, found)
	require.Nil(t, cred)
}

func TestFlush(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", &CachedCredentials{LoginSuccess: true, ClientID: "1"})
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "1", cred.ClientID)
	require.True(t, cred.LoginSuccess)

	c.Flush()
	cred, found = c.Get("key")
	require.False(t, found)
	require.Nil(t, cred)
}
