package lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	logger.Init("Debug")
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10, 100)

		_, ok := c.Get(Key(1))
		require.False(t, ok)

		_, ok = c.Get(Key(10))
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(2, 0)

		success := c.Set(100, "aaa")
		require.True(t, success)

		success = c.Set(200, "bbb")
		require.True(t, success)

		val, ok := c.Get(1)
		require.True(t, ok)
		require.Equal(t, "bbb", val[0])

		val, ok = c.Get(101)
		require.True(t, ok)
		require.Equal(t, "aaa", val[1])

		success = c.Set(300, "ccc")
		require.True(t, success)

		val, ok = c.Get(1)
		require.True(t, ok)
		require.Equal(t, "ccc", val[0])

		val, ok = c.Get(301)
		require.True(t, ok)
		require.Equal(t, 2, len(val))
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(100, 0)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(i), strconv.Itoa(i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(rand.Intn(1_000_000)))
		}
	}()

	wg.Wait()

	require.Nil(t, nil)
}
