//go:build windows

package sysstats

import (
	"errors"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

func GetDiskUsage() ([]*api.DiskUsage, error) {
	err := "Not release for OS Windows"
	logger.Log.WithFields(logrus.Fields{
		"file": "disk_usage_windows.go",
		"func": "GetDiskUsage()",
	}).Error(err)
	return nil, errors.New("Not release for OS Windows")
}
