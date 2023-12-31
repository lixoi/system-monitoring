package config

import (
	"fmt"
	"io/ioutil" //nolint:all
	"strconv"

	"github.com/valyala/fastjson"
)

type Config struct {
	Server     ServerConf
	DumpFields DumpConf
	LogLevel   string
}

type ServerConf struct {
	Port     string
	Capacity int
}

type DumpConf struct {
	ConnectStats      bool
	DiskStats         bool
	DiskUsage         bool
	LoadAverage       bool
	LoadCPU           bool
	LoadDisks         bool
	NetworkTopTalkers TopTalkersConfig
}

type TopTalkersConfig struct {
	Enable bool
	TCP    bool
	UDP    bool
	ICMP   bool
}

func NewConfig(fpath string) (c Config, err error) { //nolint:all
	// filename is the JSON file to read
	config, err := ioutil.ReadFile(fpath)
	if err != nil {
		return
	}

	v, err := fastjson.ParseBytes(config)
	if err != nil {
		return
	}
	// parse Logger parameters
	// you can set levels: Debug, Warning
	if !v.Exists("Logger") {
		err = fmt.Errorf("not init Logger in %s", fpath)
		return
	}
	vv := v.Get("Logger")
	if !vv.Exists("Level") {
		err = fmt.Errorf("not init Log level config in %s", fpath)
		return
	}
	c.LogLevel = string(vv.Get("Level").GetStringBytes())
	// parse Server parameters
	if !v.Exists("Server") {
		err = fmt.Errorf("not init Server in %s", fpath)
		return
	}
	vv = v.Get("Server")
	if !vv.Exists("Port") || !vv.Exists("Capacity") || !vv.Exists("Timeout") {
		err = fmt.Errorf("not init Server config in %s", fpath)
		return
	}
	c.Server.Port = string(vv.Get("Port").GetStringBytes())
	c.Server.Capacity, err = strconv.Atoi(string(vv.Get("Port").GetStringBytes()))
	if err != nil {
		err = fmt.Errorf("not init Capacity parameters of Server in %s", fpath)
		return
	}
	// parse DumpFields parameters
	if !v.Exists("DumpFields") {
		err = fmt.Errorf("not init DumpFields config in %s", fpath)
		return
	}
	vv = v.Get("DumpFields")
	if !vv.Exists("ConnectStats") || !vv.Exists("DiskStats") ||
		!vv.Exists("DiskUsage") || !vv.Exists("LoadAverage") ||
		!vv.Exists("LoadCPU") || !vv.Exists("LoadDisks") {
		err = fmt.Errorf("not init parameters of DumpFields in %s", fpath)
		return
	}
	if c.DumpFields.ConnectStats, err = strconv.ParseBool(string(vv.Get("ConnectStats").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.DiskStats, err = strconv.ParseBool(string(vv.Get("DiskStats").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.DiskUsage, err = strconv.ParseBool(string(vv.Get("DiskUsage").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.LoadAverage, err = strconv.ParseBool(string(vv.Get("LoadAverage").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.LoadCPU, err = strconv.ParseBool(string(vv.Get("LoadCPU").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.LoadDisks, err = strconv.ParseBool(string(vv.Get("LoadDisks").GetStringBytes())); err != nil {
		return
	}
	// parse TopTalkersConfig parameters
	if !vv.Exists("NetworkTopTalkers") {
		err = fmt.Errorf("not init NetworkTopTalkers config in %s", fpath)
		return
	}
	vvv := vv.Get("NetworkTopTalkers")
	if !vvv.Exists("Enable") {
		err = fmt.Errorf("not init parameters of NetworkTopTalkers in %s", fpath)
		return
	}
	if c.DumpFields.NetworkTopTalkers.Enable,
		err = strconv.ParseBool(string(vvv.Get("Enable").GetStringBytes())); err != nil {
		return
	}
	if !c.DumpFields.NetworkTopTalkers.Enable {
		return
	}
	if !vvv.Exists("TCP") || !vvv.Exists("UDP") || !vvv.Exists("ICMP") {
		err = fmt.Errorf("not init parameters of NetworkTopTalkers in %s", fpath)
		return
	}
	if c.DumpFields.NetworkTopTalkers.TCP, err = strconv.ParseBool(string(vvv.Get("TCP").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.NetworkTopTalkers.UDP, err = strconv.ParseBool(string(vvv.Get("UDP").GetStringBytes())); err != nil {
		return
	}
	if c.DumpFields.NetworkTopTalkers.ICMP, err = strconv.ParseBool(string(vvv.Get("ICMP").GetStringBytes())); err != nil {
		return
	}

	return
}
