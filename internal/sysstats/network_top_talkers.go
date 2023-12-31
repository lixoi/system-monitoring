package sysstats

import (
	"errors"
	"net"
	"sort"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/lixoi/system_stats_daemon/config"
	cache "github.com/lixoi/system_stats_daemon/internal/memory/lru_cache"
	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const (
	cTCP  = "TCP"
	cUDP  = "UDP"
	cICMP = "ICMP"
)

type NetStats struct {
	TimeStamp time.Time
	Length    uint32
	Type      string
	Protocol  string
	SrcIP     string
	SrcPort   string
	DstIP     string
	DstPort   string
}

type NetworkSniffer struct {
	Buffer cache.Cache
	config config.TopTalkersConfig
}

func NewNetworkSniffer(capacity int, timeInterval cache.Key, conf config.TopTalkersConfig) NetworkSniffer {
	return NetworkSniffer{
		Buffer: cache.NewCache(capacity, timeInterval),
		config: conf,
	}
}

func (ns *NetworkSniffer) Start() error {
	infs, _ := net.Interfaces()
	for _, f := range infs {
		if f.Name == "lo" {
			continue
		}
		if addrs, err := f.Addrs(); err == nil && len(addrs) > 0 {
			go ns.netInterfaceSniffing(f.Name)
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "network_top_talkers.go",
		"func": "Start()",
	}).Debug("")

	return nil
}

func (ns *NetworkSniffer) netInterfaceSniffing(ether string) {
	var (
		eth     layers.Ethernet
		ip4     layers.IPv4
		icmpv4  layers.ICMPv4
		tcp     layers.TCP
		udp     layers.UDP
		payload gopacket.Payload
	)
	decodedLayers := make([]gopacket.LayerType, 0, 10)
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &tcp, &udp, &icmpv4, &payload)
	_ = parser
	if handle, err := pcap.OpenLive(ether, 1600, true, pcap.BlockForever); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "network_top_talkers.go",
			"func": "netInterfaceSniffing()",
		}).Error(err.Error())
		return
	} else if err := handle.SetBPFFilter("tcp or udp or icmp"); err != nil { // optional
		logger.Log.WithFields(logrus.Fields{
			"file": "network_top_talkers.go",
			"func": "netInterfaceSniffing()",
		}).Error(err.Error())
		return
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			stat := NetStats{}
			stat.TimeStamp = packet.Metadata().Timestamp
			stat.Length = uint32(packet.Metadata().Length)
			parser.DecodeLayers(packet.Data(), &decodedLayers)
			if len(decodedLayers) < 3 {
				continue
			}
			for _, typ := range decodedLayers {
				switch typ {
				case layers.LayerTypeIPv4:
					stat.SrcIP, stat.DstIP = ip4.SrcIP.String(), ip4.DstIP.String()
				case layers.LayerTypeICMPv4:
					stat.SrcIP, stat.DstIP = ip4.SrcIP.String(), ip4.DstIP.String()
					stat.Type = cICMP
				case layers.LayerTypeTCP:
					stat.SrcPort, stat.DstPort = tcp.SrcPort.String(), tcp.DstPort.String()
					stat.Type = cTCP
				case layers.LayerTypeUDP:
					stat.SrcPort, stat.DstPort = udp.SrcPort.String(), udp.DstPort.String()
					stat.Type = cUDP
				}
			}
			ns.Buffer.Set(cache.Key(stat.TimeStamp.UnixNano()), stat)
		}
	}
}

func (ns *NetworkSniffer) GetNetworkTopTalkers(
	interval cache.Key,
) ([]*api.TopTalkersProtocol, []*api.TopTalkersTraffic, error) {
	slice, isExist := ns.Buffer.Get(interval)
	if !isExist {
		err := "network dump is empty"
		logger.Log.WithFields(logrus.Fields{
			"file": "network_top_talkers.go",
			"func": "netInterfaceSniffing()",
		}).Error(err)
		return nil, nil, errors.New(err)
	}

	allTraffic := ns.getAllTraffic(slice)
	ttp := ns.getAllTrafficForProtocol(slice)
	for i := 0; i < len(ttp); i++ {
		ttp[i].Rate = (ttp[i].Bytes * 100) / allTraffic
	}

	sort.Slice(ttp, func(i, j int) bool {
		return ttp[i].Rate < ttp[j].Rate
	})

	ttt := ns.getAllTrafficForSourse(slice)

	logger.Log.WithFields(logrus.Fields{
		"file": "network_top_talkers.go",
		"func": "GetNetworkTopTalkers()",
	}).Debug("")

	return ttp, ttt, nil
}

func (ns *NetworkSniffer) getAllTraffic(slice []interface{}) uint32 {
	var res uint32
	for i := range slice {
		res += slice[i].(NetStats).Length
	}

	return res
}

func (ns *NetworkSniffer) getAllTrafficForProtocol(slice []interface{}) []*api.TopTalkersProtocol {
	mapTTP := make(map[string]*api.TopTalkersProtocol, 3)
	for i := range slice {
		sl := slice[i].(NetStats)
		if ns.isDisableProtocol(sl.Type) {
			continue
		}
		if v, ok := mapTTP[sl.Type]; ok {
			v.Bytes += sl.Length
			mapTTP[sl.Type] = v
		} else {
			mapTTP[sl.Type] = &api.TopTalkersProtocol{
				Protocol: sl.Type,
				Bytes:    sl.Length,
			}
		}
	}

	return maps.Values(mapTTP)
}

func (ns *NetworkSniffer) getAllTrafficForSourse(slice []interface{}) []*api.TopTalkersTraffic {
	mapTTT := make(map[string]*api.TopTalkersTraffic, 10)
	for i := range slice {
		sl := slice[i].(NetStats)
		if ns.isDisableProtocol(sl.Type) {
			continue
		}
		if v, ok := mapTTT[sl.SrcIP+sl.SrcPort]; ok {
			v.Bps += sl.Length
			mapTTT[sl.SrcIP+sl.SrcPort] = v
		} else {
			mapTTT[sl.SrcIP+sl.SrcPort] = &api.TopTalkersTraffic{
				Bps:         sl.Length,
				Distination: sl.DstIP + ":" + sl.DstPort,
				Source:      sl.SrcIP + ":" + sl.SrcPort,
				Protocol:    sl.Type,
			}
		}
	}

	return maps.Values(mapTTT)
}

func (ns *NetworkSniffer) isDisableProtocol(protocol string) bool {
	switch protocol {
	case cTCP:
		if !ns.config.TCP {
			return true
		}
	case cUDP:
		if !ns.config.UDP {
			return true
		}
	case cICMP:
		if !ns.config.UDP {
			return true
		}
	}

	return false
}
