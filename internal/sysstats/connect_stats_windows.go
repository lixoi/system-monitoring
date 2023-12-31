//go:build windows

package sysstats

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

func GetListeningSockets() ([]*api.ListeningSocket, error) {
	// Check if netstat is exists
	netstat, err := exec.LookPath("netstat")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "GetListeningSockets()",
		}).Error(err.Error())
		return nil, err
	}
	out, err := exec.Command(netstat, "-ano").Output()
	if err != nil {
		return nil, err
	}
	mapPs := make(map[uint32][]*api.ListeningSocket, 20)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserNetStat(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "connect_stats_windows.go",
				"func": "GetListeningSockets()",
			}).Error(err.Error())
			return nil, err
		}
		mapPs[state.Pid] = append(mapPs[state.Pid], state)
	}
	// Check if tasklist is exists
	tasklist, err := exec.LookPath("tasklist")
	if err != nil {
		return nil, err
	}
	out, err = exec.Command(tasklist, "/V").Output()
	if err != nil {
		return nil, err
	}
	res := make([]*api.ListeningSocket, 0, 5)
	scanner = bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserTaskList(line)
		if err != nil {
			return nil, err
		}
		if v, ok := mapPs[state.Pid]; ok {
			for i := range v {
				v[i].Command = state.Command
				v[i].User = state.User
				res = append(res, v[i])
			}
		}
	}

	return res, nil
}

func GetConnects() ([]*api.Connect, error) {
	// Check if netstat is exists
	ss, err := exec.LookPath("netstat")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "GetConnects()",
		}).Error(err.Error())
		return nil, err
	}
	out, err := exec.Command(ss, "-an").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "GetConnects()",
		}).Error(err.Error())
		return nil, err
	}
	mapSS := make(map[string]*api.Connect, 20)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserConn(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "connect_stats_windows.go",
				"func": "GetConnects()",
			}).Error(err.Error())
			return nil, err
		}
		if v, ok := mapSS[state]; ok {
			v.State = state
			v.Number += 1
			mapSS[state] = v
		} else {
			mapSS[state] = &api.Connect{State: state, Number: 1}
		}
	}

	return maps.Values(mapSS), nil
}

func parserNetStat(line string) (*api.ListeningSocket, error) {
	ls := &api.ListeningSocket{}
	fields := strings.Fields(line)
	if len(fields) < 4 {
		err := "Couldn't parse netstat because there are less than 4 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "parserNetStat()",
		}).Error(err)
		return ls, errors.New(err)
	}
	// Parse fields
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch i {
		case 0:
			ls.Protocol = field
		case 1:
			addr := strings.Split(field, ":")
			if len(addr) == 0 {
				continue
			}
			value, err := strconv.ParseUint(addr[len(addr)-1], 10, 16)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_windows.go",
					"func": "parserNetStat()",
				}).Error(err.Error())
				return &api.ListeningSocket{}, err
			}
			ls.Port = uint32(value)
		case 3:
			if len(field) == 4 {
				break
			}
			pid, err := strconv.ParseUint(field, 10, 32)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_windows.go",
					"func": "parserNetStat()",
				}).Error(err.Error())
				return &api.ListeningSocket{}, err
			}
			ls.Pid = uint32(pid)
			return ls, nil
		case 4:
			pid, err := strconv.ParseUint(field, 10, 32)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_windows.go",
					"func": "parserNetStat()",
				}).Error(err.Error())
				return &api.ListeningSocket{}, err
			}
			ls.Pid = uint32(pid)
		}
	}

	return ls, nil
}

func parserTaskList(line string) (api.ListeningSocket, error) {
	ls := api.ListeningSocket{}
	fields := strings.Fields(line)
	if len(fields) < 9 {
		err := "Couldn't parse tasklist because there aren't 9 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "parserTaskList()",
		}).Error(err)
		return ls, errors.New(err)
	}
	// Parse fields
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch i {
		case 0:
			ls.Command = field
			break
		case 1:
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_windows.go",
					"func": "parserTaskList()",
				}).Error(err.Error())
				return api.ListeningSocket{}, err
			}
			ls.Pid = uint32(value)
			break
		case 6:
			ls.User = field
			break
		}
	}

	return ls, nil
}

func parserConn(line string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		err := "Couldn't parse ss because there aren't less 4 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_windows.go",
			"func": "parserTaskList()",
		}).Error(err)
		return "", errors.New(err)
	}

	return fields[3], nil
}
