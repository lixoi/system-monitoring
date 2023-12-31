package sysstats

import (
	"testing"

	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/stretchr/testify/require"
)

func TestGetListeningSockets(t *testing.T) {
	logger.Init("Debug")
	t.Run("sympe", func(t *testing.T) {
		c, _ := GetListeningSockets()
		require.Nil(t, c)
	})
}

func TestGetConnects(t *testing.T) {
	logger.Init("Debug")
	t.Run("sympe", func(t *testing.T) {
		_, err := GetConnects()
		require.Nil(t, err)
	})
}
