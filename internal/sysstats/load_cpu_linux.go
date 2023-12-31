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

func GetLoadCPU() (*api.LoadCPU, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_cpu_linux.go",
			"func": "GetLoadCPU()",
		}).Error(err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		if res, err := parseCPUStats(line); err == nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "load_cpu_linux.go",
				"func": "GetLoadCPU()",
			}).Debug("")
			return res, nil
		}
	}

	errorNotData := "not data into file /proc/stat"
	logger.Log.WithFields(logrus.Fields{
		"file": "load_cpu_linux.go",
		"func": "GetLoadCPU()",
	}).Error(errorNotData)

	return nil, errors.New(errorNotData)
}

func parseCPUStats(stats string) (*api.LoadCPU, error) {
	loadCPU := &api.LoadCPU{}
	fields := strings.Fields(stats)
	if strings.Compare(fields[0], "cpu") != 0 {
		errorNotData := "not data into file /proc/stat"
		logger.Log.WithFields(logrus.Fields{
			"file": "load_cpu_linux.go",
			"func": "parseCPUStats()",
		}).Error(errorNotData)
		return nil, errors.New(errorNotData)
	}
	for i := 1; i < len(fields); i++ {
		stat, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "load_cpu_linux.go",
				"func": "parseCPUStats()",
			}).Error(err.Error())
			return nil, err
		}
		switch i {
		case 1:
			loadCPU.UserMode = stat
		case 3:
			loadCPU.SystemMode = stat
		case 4:
			loadCPU.Idle = stat
		}
	}

	return loadCPU, nil
}
