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

func GetListeningSockets() ([]*api.ListeningSocket, error) {
	// Check if netstat is exists
	netstat, err := exec.LookPath("netstat")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err.Error())
		return nil, err
	}
	out, err := exec.Command(netstat, "-lntup").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err.Error())
		return nil, err
	}
	mapPs := make(map[uint32][]*api.ListeningSocket, 20)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserNetStat(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "connect_stats_linux.go",
				"func": "GetListeningSockets()",
			}).Error(err.Error())
			return nil, err
		}
		mapPs[state.Pid] = append(mapPs[state.Pid], state)
	}
	// Check if ps is exists
	ps, err := exec.LookPath("ps")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err.Error())
		return nil, err
	}
	out, err = exec.Command(ps, "-aux").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err.Error())
		return nil, err
	}
	res := make([]*api.ListeningSocket, 0, 5)
	scanner = bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserPS(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "connect_stats_linux.go",
				"func": "GetListeningSockets()",
			}).Error(err.Error())
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

	logger.Log.WithFields(logrus.Fields{
		"file": "connect_stats_linux.go",
		"func": "GetListeningSockets()",
	}).Debug("")

	return res, nil
}

func GetConnects() ([]*api.Connect, error) {
	// Check if ss is exists
	ss, err := exec.LookPath("ss")
	if err != nil {
		return nil, err
	}
	out, err := exec.Command(ss, "-ta").Output()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetConnects()",
		}).Error(err.Error())
		return nil, err
	}
	mapSS := make(map[string]*api.Connect, 20)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	// Filter the header
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		state, err := parserSS(line)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "connect_stats_linux.go",
				"func": "GetConnects()",
			}).Error(err.Error())
			return nil, err
		}
		if v, ok := mapSS[state]; ok {
			v.State = state
			v.Number++
			mapSS[state] = v
		} else {
			mapSS[state] = &api.Connect{State: state, Number: 1}
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "connect_stats_linux.go",
		"func": "GetConnects()",
	}).Debug("")

	return maps.Values(mapSS), nil
}

func parserNetStat(line string) (*api.ListeningSocket, error) {
	ls := &api.ListeningSocket{}
	fields := strings.Fields(line)
	if len(fields) < 6 {
		err := "couldn't parse netstat because there are less than 6 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
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
		case 3:
			addr := strings.Split(field, ":")
			if len(addr) == 0 {
				continue
			}
			value, err := strconv.ParseUint(addr[len(addr)-1], 10, 16)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_linux.go",
					"func": "parserNetStat()",
				}).Error(err)
				return &api.ListeningSocket{}, err
			}
			ls.Port = uint32(value)
		case 5:
			if field == "LISTEN" {
				break
			}
			program := strings.Split(field, "/")
			if len(program) == 0 {
				continue
			}
			pid, err := strconv.ParseUint(program[0], 10, 32)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_linux.go",
					"func": "parserNetStat()",
				}).Error(err)
				return &api.ListeningSocket{}, err
			}
			ls.Pid = uint32(pid)
			return ls, nil
		case 6:
			program := strings.Split(field, "/")
			if len(program) == 0 {
				continue
			}
			pid, err := strconv.ParseUint(program[0], 10, 32)
			if err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "connect_stats_linux.go",
					"func": "parserNetStat()",
				}).Error(err)
				return &api.ListeningSocket{}, err
			}
			ls.Pid = uint32(pid)
		}
	}

	return ls, nil
}

func parserPS(line string) (api.ListeningSocket, error) {
	ls := api.ListeningSocket{}
	fields := strings.Fields(line)
	if len(fields) < 11 {
		err := "couldn't parse ps because there aren't 11 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err)
		return ls, errors.New(err) //nolint:all
	}
	// Parse fields
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch i {
		case 0:
			ls.User = field
		case 1:
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return api.ListeningSocket{}, err
			}
			ls.Pid = uint32(value)
		case 10:
			ls.Command = field
		default:
			if i > 10 {
				ls.Command += " " + field
			}
		}
	}

	return ls, nil //nolint:all
}

func parserSS(line string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		err := "couldn't parse ss because there aren't less 5 fields"
		logger.Log.WithFields(logrus.Fields{
			"file": "connect_stats_linux.go",
			"func": "GetListeningSockets()",
		}).Error(err)
		return "", errors.New(err)
	}

	return fields[0], nil
}
