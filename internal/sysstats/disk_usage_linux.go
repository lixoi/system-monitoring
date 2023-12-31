//go:build linux

package sysstats

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func GetDiskUsage() ([]*api.DiskUsage, error) {
	// Check if df is exists
	df, err := exec.LookPath("df")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_usage_linux.go",
			"func": "GetDiskUsage()",
		}).Error(err.Error())
		return nil, err
	}
	// Use Mb
	out, err := exec.Command(df, "-k").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_usage_linux.go",
			"func": "GetDiskUsage()",
		}).Error(err.Error())
		return nil, err
	}
	disk := make(map[string]*api.DiskUsage, 5)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		diskUsage, err := parserDF(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "disk_usage_linux.go",
				"func": "GetDiskUsage()",
			}).Error(err.Error())
			return nil, err
		}
		if v, ok := disk[diskUsage.FileSystem]; ok {
			v.Use += diskUsage.Use
			v.Used += diskUsage.Used
			disk[diskUsage.FileSystem] = v
		} else {
			disk[diskUsage.FileSystem] = diskUsage
		}
	}
	// Use Inode
	out, err = exec.Command(df, "-i").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "disk_usage_linux.go",
			"func": "GetDiskUsage()",
		}).Error(err.Error())
		return nil, err
	}
	scanner = bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		diskUsage, err := parserDF(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "disk_usage_linux.go",
				"func": "GetDiskUsage()",
			}).Error(err.Error())
			return nil, err
		}
		if v, ok := disk[diskUsage.FileSystem]; ok {
			v.Iuse += diskUsage.Use
			v.Iused += diskUsage.Used
			disk[diskUsage.FileSystem] = v
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "disk_usage_linux.go",
		"func": "GetDiskUsage()",
	}).Debug("")

	return maps.Values(disk), nil
}

func parserDF(line string) (*api.DiskUsage, error) {
	diskUsage := &api.DiskUsage{}
	fields := strings.Fields(line)
	if len(fields) != 6 {
		return diskUsage, errors.New("couldn't parse disk usage because there aren't 6 fields")
	}
	// Parse fields
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch i {
		case 0:
			diskUsage.FileSystem = field
		case 2:
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return &api.DiskUsage{}, err
			}
			diskUsage.Used = value
		case 4:
			if last := len(field) - 1; last >= 0 && (field[last] == '%' || field[last] == '-') {
				field = field[:last]
			}
			if field == "" {
				continue
			}
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return &api.DiskUsage{}, err
			}
			diskUsage.Use = value
		}
	}

	return diskUsage, nil
}
