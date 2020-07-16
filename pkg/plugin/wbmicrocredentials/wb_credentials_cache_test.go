package wbmicrocredentials

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetFound(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", NewCachedCredentials("client", "user", true))
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "client", cred.ClientID)
	require.Equal(t, "user", cred.UserID)
	require.True(t, cred.LoginSuccess)
}

func TestGetNotFound(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", NewCachedCredentials("client", "user", true))
	cred, found := c.Get("key2")
	require.False(t, found)
	require.Nil(t, cred)
}

func TestReplace(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", NewCachedCredentials("client", "user", true))
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "client", cred.ClientID)
	require.Equal(t, "user", cred.UserID)
	require.True(t, cred.LoginSuccess)
	c.Put("key", NewCachedCredentials("client2", "user2", false))
	cred, found = c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "client2", cred.ClientID)
	require.Equal(t, "user2", cred.UserID)
	require.False(t, cred.LoginSuccess)
}

func TestExpiration(t *testing.T) {
	c := NewCredentialsCache(1*time.Second, 5*time.Minute)
	c.Put("key", NewCachedCredentials("client", "user", true))
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "client", cred.ClientID)
	require.Equal(t, "user", cred.UserID)
	require.True(t, cred.LoginSuccess)

	// wait until the record expires
	time.Sleep(1 * time.Second)
	cred, found = c.Get("key")
	require.False(t, found)
	require.Nil(t, cred)
}

func TestFlush(t *testing.T) {
	c := NewCredentialsCache(5*time.Minute, 5*time.Minute)
	c.Put("key", NewCachedCredentials("client", "user", true))
	cred, found := c.Get("key")
	require.True(t, found)
	require.NotNil(t, cred)
	require.Equal(t, "client", cred.ClientID)
	require.Equal(t, "user", cred.UserID)
	require.True(t, cred.LoginSuccess)

	c.Flush()
	cred, found = c.Get("key")
	require.False(t, found)
	require.Nil(t, cred)
}
