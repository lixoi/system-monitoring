//go:build linux

package sysstats

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

type LoadDisk struct {
	DiskDevice string  `json:"disk_device"` //nolint:all
	Tps        float64 `json:"tps"`         //nolint:all
	KbRps      float64 `json:"kB_read/s"`   //nolint:all
	KbWps      float64 `json:"kB_wrtn/s"`   //nolint:all
}

type Iostat struct {
	Sysstat struct {
		Hosts []struct {
			Statistics []struct {
				Disk []LoadDisk
			}
		}
	}
}

func GetLoadDisk() ([]*api.LoadDisk, error) {
	loadDisk := make([]*api.LoadDisk, 0, 5)
	// Check if iostat is exists
	iostat, err := exec.LookPath("iostat")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_disk_linux.go",
			"func": "GetLoadDisk()",
		}).Error(err.Error())
		return nil, err
	}

	out, err := exec.Command(iostat, "-d", "-k", "-o", "JSON").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_disk_linux.go",
			"func": "GetLoadDisk()",
		}).Error(err.Error())
		return nil, err
	}

	iostatOut := Iostat{}
	if err = json.Unmarshal(out, &iostatOut); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "load_disk_linux.go",
			"func": "GetLoadDisk()",
		}).Error(err.Error())
		return nil, err
	}

	if len(iostatOut.Sysstat.Hosts) == 0 || len(iostatOut.Sysstat.Hosts[0].Statistics) == 0 {
		errDev := "not devices in out of iostat command"
		logger.Log.WithFields(logrus.Fields{
			"file": "load_disk_linux.go",
			"func": "GetLoadDisk()",
		}).Error(errDev)
		return nil, errors.New(errDev)
	}

	disks := iostatOut.Sysstat.Hosts[0].Statistics[0].Disk
	for i := range disks {
		if strings.HasPrefix(disks[i].DiskDevice, "loop") {
			continue
		}
		loadDisk = append(loadDisk, &api.LoadDisk{
			DiskDevice: disks[i].DiskDevice,
			Tps:        disks[i].Tps,
			KbRps:      disks[i].KbRps,
			KbWps:      disks[i].KbWps,
			KbPs:       disks[i].KbRps + disks[i].KbWps,
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "load_disk_linux.go",
		"func": "GetLoadDisk()",
	}).Debug("")

	return loadDisk, nil
}
