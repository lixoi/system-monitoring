package daemon

import (
	"context"
	"errors"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/lixoi/system_stats_daemon/config"
	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/internal/server/grpc/validate"
	systemdump "github.com/lixoi/system_stats_daemon/internal/system_dump"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var configFile string

type Service struct {
	api.UnimplementedSystemStatisticsServer
	cache          systemdump.CacheSysStatDumps
	requestCounter []int
	conf           config.ServerConf
}

func NewService(conf config.Config) *Service {
	return &Service{
		cache:          systemdump.NewCacheSysStatDumps(conf),
		requestCounter: make([]int, 0, 10),
		conf:           conf.Server,
	}
}

func (s *Service) StreamSystemDump(
	in *api.GetSystemDumpRequest,
	stream api.SystemStatistics_StreamSystemDumpServer,
) error {
	s.addRequest(int(in.GetM()))
	defer s.delRequest(int(in.GetM()))

	logger.Log.WithFields(logrus.Fields{
		"file": "grpc_server.go",
		"func": "StreamSystemDump()",
	}).Debug("call " + strconv.Itoa(len(s.requestCounter)) + " GRPC func")

	ticker := time.NewTicker(time.Duration(in.GetN()) * time.Second)
	for { //nolint:all
		select {
		case <-ticker.C:
			dump := s.cache.GetSysStatDumpOver(in.GetM())
			if err := stream.Send(&api.GetSystemDumpResponse{SystemDump: dump}); err != nil {
				logger.Log.WithFields(logrus.Fields{
					"file": "grpc_server.go",
				}).Error(err.Error())
				return err
			}
		}
	}

	// return nil
}

func (s *Service) GetSystemDump(
	ctx context.Context,
	in *api.GetSystemDumpRequest,
) (*api.GetSystemDumpResponse, error) {
	_ = ctx
	s.addRequest(int(in.GetM()))
	defer s.delRequest(int(in.GetM()))

	dump := s.cache.GetSysStatDumpOver(in.GetM())

	return &api.GetSystemDumpResponse{SystemDump: dump}, nil
}

func (s *Service) addRequest(m int) {
	s.requestCounter = append(s.requestCounter, m)
	if m > s.cache.Buffer.GetCapacity() {
		s.cache.Buffer.ResizeCacheOfCap(m)

		logger.Log.WithFields(logrus.Fields{
			"file": "grpc_server.go",
		}).Debug("reallocate cache, current size = " + strconv.Itoa(m))
	}
}

func (s *Service) delRequest(m int) {
	for i := range s.requestCounter {
		if s.requestCounter[i] == m {
			s.requestCounter = append(s.requestCounter[:i], s.requestCounter[i+1:]...)
			break
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"file": "grpc_server.go",
	}).Debug("there is " + strconv.Itoa(len(s.requestCounter)) + " active GRPC funcs")

	if len(s.requestCounter) == 0 {
		s.cache.Buffer.ResizeCacheOfCap(s.conf.Capacity)
		return
	}
	if m == s.cache.Buffer.GetCapacity() {
		sort.Slice(s.requestCounter, func(i, j int) bool {
			return s.requestCounter[i] > s.requestCounter[j]
		})
		s.cache.Buffer.ResizeCacheOfCap(s.requestCounter[0])
	}
}

func (s *Service) Start(port string) error {
	if s.cache.StartDump() != nil {
		return errors.New("error init cache")
	}
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "grpc_server.go",
		}).Error(err.Error())
		return err
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			validate.UnaryServerRequestValidatorInterceptor(validate.Req),
		),
		grpc.ChainStreamInterceptor(
			validate.StreamServerRequestValidatorInterceptor(validate.Req),
		),
	)
	api.RegisterSystemStatisticsServer(srv, s)
	return srv.Serve(l)
}

var GrpcServerCmd = &cobra.Command{
	Use:   "grpc_server",
	Short: "Run grpc server",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.NewConfig(configFile)
		if err != nil {
			return
		}
		logger.Init(config.LogLevel)
		server := NewService(config)
		server.Start(config.Server.Port)
	},
}

func init() {
	GrpcServerCmd.Flags().StringVar(&configFile, "configFile", "config.json", "path to configuration file for Server")
}
