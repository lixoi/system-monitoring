//go:build linux

package sysstats

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

func GetDiskStats() (*api.DiskStats, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_stats_linux.go",
			"func": "GetDiskStats()",
		}).Error(err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "loop") {
			logger.Log.WithFields(logrus.Fields{
				"file": "disk_stats_linux.go",
				"func": "GetDiskStats()",
			}).Debug("")
			return parseDiskStats(line)
		}
	}

	errMsg := "not data"
	logger.Log.WithFields(logrus.Fields{
		"file": "disk_stats_linux.go",
		"func": "GetDiskStats()",
	}).Error(err)
	return nil, errors.New(errMsg)
}

func parseDiskStats(stats string) (*api.DiskStats, error) {
	diskStats := &api.DiskStats{}
	fields := strings.Fields(stats)
	if len(fields) < 14 {
		err := "couldn't parse /proc/diskstats because there are less than 14 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_stats_linux.go",
			"func": "parseDiskStats()",
		}).Error(err)
		return nil, errors.New(err)
	}

	ioTime, err := strconv.ParseFloat(fields[12], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_stats_linux.go",
			"func": "parseDiskStats()",
		}).Error(err.Error())
		return nil, err
	}
	diskStats.IoTime = ioTime

	ioInProgress, err := strconv.ParseFloat(fields[11], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_stats_linux.go",
			"func": "parseDiskStats()",
		}).Error(err.Error())
		return nil, err
	}
	diskStats.IoInProgress = ioInProgress

	weightedIO, err := strconv.ParseFloat(fields[13], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_stats_linux.go",
			"func": "parseDiskStats()",
		}).Error(err.Error())
		return nil, err
	}
	diskStats.WeightedIo = weightedIO

	return diskStats, nil
}
