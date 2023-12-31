package systemdump

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/lixoi/system_stats_daemon/config"
	lrucache "github.com/lixoi/system_stats_daemon/internal/memory/lru_cache"
	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	sysstats "github.com/lixoi/system_stats_daemon/internal/sysstats"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
)

type CacheSysStatDumps struct {
	mu     sync.Mutex
	Buffer lrucache.Cache
	config config.DumpConf
}

func NewCacheSysStatDumps(conf config.Config) CacheSysStatDumps {
	return CacheSysStatDumps{
		Buffer: lrucache.NewCache(conf.Server.Capacity, 0),
		config: conf.DumpFields,
	}
}

func (cssd *CacheSysStatDumps) StartDump() error { //nolint:all
	logger.Log.WithFields(logrus.Fields{
		"file": "sniffer.go",
		"func": "Start()",
	}).Debug("start system dump sniffer")
	cicle, _ := time.ParseDuration("1s")
	ticker := time.NewTicker(1 * time.Second)
	ns := sysstats.NetworkSniffer{}

	if cssd.config.NetworkTopTalkers.Enable {
		interval, _ := time.ParseDuration("1.5s")
		ns = sysstats.NewNetworkSniffer(
			int(interval.Milliseconds()),
			lrucache.Key(interval.Nanoseconds()),
			cssd.config.NetworkTopTalkers,
		)
		ns.Start()
	}
	go func() {
		for { //nolint:all
			select {
			case <-ticker.C:
				dump := api.SystemDump{}
				// init system stats
				if cssd.config.LoadAverage {
					dump.LA, _ = sysstats.GetLoadAvg()
				}
				if cssd.config.LoadCPU {
					dump.LC, _ = sysstats.GetLoadCPU()
				}
				if cssd.config.DiskStats {
					dump.DS, _ = sysstats.GetDiskStats()
				}
				if cssd.config.LoadDisks {
					dump.LD, _ = sysstats.GetLoadDisk()
				}
				if cssd.config.DiskUsage {
					dump.DU, _ = sysstats.GetDiskUsage()
				}
				if cssd.config.ConnectStats {
					dump.CS = &api.ConnectStats{}
					dump.CS.Conn, _ = sysstats.GetConnects()
					dump.CS.Ls, _ = sysstats.GetListeningSockets()
				}
				dump.TT = &api.TopTalkers{}
				if cssd.config.NetworkTopTalkers.Enable {
					dump.TT.Ttp, dump.TT.Ttt, _ = ns.GetNetworkTopTalkers(lrucache.Key(cicle.Nanoseconds()))
				}
				// save in cache
				cssd.mu.Lock()
				cssd.Buffer.Set(lrucache.Key(time.Now().UnixNano()), dump) //nolint:all
				cssd.mu.Unlock()
			}
		}
	}()

	return nil
}

func (cssd *CacheSysStatDumps) ChangeSizeCache(capacity int) {
	cssd.Buffer.ResizeCacheOfCap(capacity)
}

func (cssd *CacheSysStatDumps) GetSysStatDumpOver(m uint32) *api.SystemDump {
	logger.Log.WithFields(logrus.Fields{
		"file": "snigger.go",
		"func": "GetSysStatDumpOver()",
	}).Debug("collect dump to send to client")

	cssd.mu.Lock()
	defer cssd.mu.Unlock()

	interval, _ := time.ParseDuration(strconv.FormatUint(uint64(m), 10) + "s")
	slice, ok := cssd.Buffer.Get(lrucache.Key(interval.Nanoseconds()))
	if !ok || len(slice) == 0 {
		return nil
	}
	if len(slice) < int(m) {
		m = uint32(len(slice))
	}

	res := cssd.copySystemDump(slice[0].(api.SystemDump)) //nolint:all

	for idx := range slice[1:] {
		dump := slice[idx].(api.SystemDump) //nolint:all
		// load disk
		for j := 0; j < len(dump.LD) && len(dump.LD) == len(res.LD); j++ {
			res.LD[j].Tps += dump.LD[j].Tps
			res.LD[j].KbPs += dump.LD[j].KbPs
			res.LD[j].KbRps += dump.LD[j].KbRps
			res.LD[j].KbWps += dump.LD[j].KbWps
		}
		// top talkers protocol (TTP)
		for j := 0; j < len(dump.TT.Ttp); j++ {
			if index := sort.Search(len(res.TT.Ttp), func(i int) bool {
				return res.TT.Ttp[i].Protocol == dump.TT.Ttp[j].Protocol
			}); index == len(res.TT.Ttp) {
				res.TT.Ttp = append(res.TT.Ttp, &api.TopTalkersProtocol{
					Protocol: dump.TT.Ttp[j].Protocol,
					Bytes:    dump.TT.Ttp[j].Bytes,
					Rate:     dump.TT.Ttp[j].Rate,
				})
			} else {
				res.TT.Ttp[index].Bytes += dump.TT.Ttp[j].Bytes
				res.TT.Ttp[index].Rate += dump.TT.Ttp[j].Rate
			}
		}
		// top talkers traffic (TTT)
		for j := 0; j < len(dump.TT.Ttt); j++ {
			if index := sort.Search(len(res.TT.Ttt), func(i int) bool {
				return res.TT.Ttt[i].Source == dump.TT.Ttt[j].Source
			}); index == len(res.TT.Ttt) {
				res.TT.Ttt = append(res.TT.Ttt, &api.TopTalkersTraffic{
					Source:      dump.TT.Ttt[j].Source,
					Distination: dump.TT.Ttt[j].Distination,
					Protocol:    dump.TT.Ttt[j].Protocol,
					Bps:         dump.TT.Ttt[j].Bps,
				})
			} else {
				res.TT.Ttt[index].Bps += dump.TT.Ttt[j].Bps
			}
		}
	}
	// calculate average
	for i := 0; i < len(res.LD); i++ { // DL
		res.LD[i].Tps /= float64(m)
		res.LD[i].KbPs /= float64(m)
		res.LD[i].KbRps /= float64(m)
		res.LD[i].KbWps /= float64(m)
	}
	for i := 0; i < len(res.TT.Ttp); i++ { // TTP
		res.TT.Ttp[i].Bytes /= m
		res.TT.Ttp[i].Rate /= m
	}
	for i := 0; i < len(res.TT.Ttt); i++ { // TTT
		res.TT.Ttt[i].Bps /= m
	}

	return res
}

func (cssd *CacheSysStatDumps) copySystemDump(sysDump api.SystemDump) *api.SystemDump { //nolint:all
	res := &api.SystemDump{}
	res.LA = &api.LoadAverage{
		AvgOneMin:     sysDump.LA.AvgOneMin,
		AvgFiveMin:    sysDump.LA.AvgFiveMin,
		AvgFifteenMin: sysDump.LA.AvgFifteenMin,
	}
	res.LC = &api.LoadCPU{
		UserMode:   sysDump.LC.UserMode,
		SystemMode: sysDump.LC.SystemMode,
		Idle:       sysDump.LC.Idle,
	}
	res.LD = make([]*api.LoadDisk, len(sysDump.LD))
	for i := range res.LD {
		res.LD[i] = &api.LoadDisk{
			DiskDevice: sysDump.LD[i].DiskDevice,
			Tps:        sysDump.LD[i].Tps,
			KbRps:      sysDump.LD[i].KbRps,
			KbWps:      sysDump.LD[i].KbWps,
			KbPs:       sysDump.LD[i].KbPs,
		}
	}
	res.DU = make([]*api.DiskUsage, len(sysDump.DU))
	for i := range res.DU {
		res.DU[i] = &api.DiskUsage{
			FileSystem: sysDump.DU[i].FileSystem,
			Used:       sysDump.DU[i].Used,
			Use:        sysDump.DU[i].Use,
			Iused:      sysDump.DU[i].Iused,
			Iuse:       sysDump.DU[i].Iuse,
		}
	}
	res.TT = &api.TopTalkers{}
	res.TT.Ttp = make([]*api.TopTalkersProtocol, len(sysDump.TT.Ttp))
	for i := range res.TT.Ttp {
		res.TT.Ttp[i] = &api.TopTalkersProtocol{
			Protocol: sysDump.TT.Ttp[i].Protocol,
			Bytes:    sysDump.TT.Ttp[i].Bytes,
			Rate:     sysDump.TT.Ttp[i].Rate,
		}
	}
	res.TT.Ttt = make([]*api.TopTalkersTraffic, len(sysDump.TT.Ttt))
	for i := range res.TT.Ttt {
		res.TT.Ttt[i] = &api.TopTalkersTraffic{
			Source:      sysDump.TT.Ttt[i].Source,
			Distination: sysDump.TT.Ttt[i].Distination,
			Protocol:    sysDump.TT.Ttt[i].Protocol,
			Bps:         sysDump.TT.Ttt[i].Bps,
		}
	}
	res.CS = &api.ConnectStats{}
	res.CS.Ls = make([]*api.ListeningSocket, len(sysDump.CS.Ls))
	for i := range res.CS.Ls {
		res.CS.Ls[i] = &api.ListeningSocket{
			Protocol: sysDump.CS.Ls[i].Protocol,
			Port:     sysDump.CS.Ls[i].Port,
			Pid:      sysDump.CS.Ls[i].Pid,
			User:     sysDump.CS.Ls[i].User,
			Command:  sysDump.CS.Ls[i].Command,
		}
	}
	res.CS.Conn = make([]*api.Connect, len(sysDump.CS.Conn))
	for i := range res.CS.Conn {
		res.CS.Conn[i] = &api.Connect{
			State:  sysDump.CS.Conn[i].State,
			Number: sysDump.CS.Conn[i].Number,
		}
	}

	return res
}
