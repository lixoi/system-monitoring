package sysstats

import (
	"testing"

	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/stretchr/testify/require"
)

func TestGetDiskStats(t *testing.T) {
	logger.Init("Debug")
	t.Run("sympe", func(t *testing.T) {
		c, err := GetDiskStats()
		require.Nil(t, err)
		require.True(t, c.IoTime > 0)
	})
}
