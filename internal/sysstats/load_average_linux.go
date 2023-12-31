//go:build linux

package sysstats

import (
	"io/ioutil" //nolint:all
	"strconv"
	"strings"
	"syscall"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

const (
	siLoadShift = 16
)

func GetLoadAvg() (*api.LoadAverage, error) {
	var sysInfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysInfo)
	if err == nil {
		return &api.LoadAverage{
			AvgOneMin:     float64(sysInfo.Loads[0]) / float64(1<<siLoadShift),
			AvgFiveMin:    float64(sysInfo.Loads[1]) / float64(1<<siLoadShift),
			AvgFifteenMin: float64(sysInfo.Loads[2]) / float64(1<<siLoadShift),
		}, nil
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "load_average_linux.go",
		"func": "GetLoadAvg()",
	}).Debug("")

	return getLoadAvgFromFile()
}

func getLoadAvgFromFile() (*api.LoadAverage, error) {
	file, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_average_linux.go",
			"func": "getLoadAvgFromFile()",
		}).Error(err.Error())
		return nil, err
	}

	loadAvg := &api.LoadAverage{}
	fields := strings.Fields(string(file))
	loadAvg.AvgOneMin, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_average_linux.go",
			"func": "getLoadAvgFromFile()",
		}).Error(err.Error())
		return nil, err
	}
	loadAvg.AvgFiveMin, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_average_linux.go",
			"func": "getLoadAvgFromFile()",
		}).Error(err.Error())
		return nil, err
	}
	loadAvg.AvgFifteenMin, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_average_linux.go",
			"func": "getLoadAvgFromFile()",
		}).Error(err.Error())
		return nil, err
	}

	return loadAvg, nil
}
