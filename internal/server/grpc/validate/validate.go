package validate

import (
	"context"
	"errors"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"github.com/lixoi/system_stats_daemon/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator func(req interface{}) error

type recvValidator struct {
	grpc.ServerStream
	validFunc Validator
	info      *grpc.StreamServerInfo
}

func (s *recvValidator) RecvMsg(m interface{}) error {
	if err := s.validFunc(m); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "validate.go",
			"func": "RecvMsg()",
		}).Error(err.Error())
		return status.Errorf(
			codes.InvalidArgument,
			"%s is rejected by validate. Error: %v",
			s.info.FullMethod, err)
	}
	if err := s.ServerStream.RecvMsg(m); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"file": "validate.go",
			"func": "RecvMsg()",
		}).Error(err.Error())
		return err
	}
	return nil
}

func UnaryServerRequestValidatorInterceptor(validator Validator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := validator(req); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file": "validate.go",
				"func": "UnaryServerRequestValidatorInterceptor()",
			}).Error(err.Error())
			return nil, status.Errorf(
				codes.InvalidArgument,
				"%s is rejected by validates. Error: %v",
				info.FullMethod,
				err)
		}
		return handler(ctx, req)
	}
}

func StreamServerRequestValidatorInterceptor(validator Validator) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		valid := &recvValidator{
			ServerStream: ss,
			validFunc:    validator,
			info:         info,
		}
		return handler(srv, valid)
	}
}

func Req(req interface{}) error {
	switch r := req.(type) {
	case *api.GetSystemDumpRequest:
		if r.GetM() < 0 || r.GetN() < 0 { //nolint:all
			err := "there are not current parameters M and N"
			logger.Log.WithFields(logrus.Fields{
				"file": "validate.go",
				"func": "Req()",
			}).Error(err)
			return errors.New(err)
		}
	default:
		logger.Log.WithFields(logrus.Fields{
			"file": "validate.go",
			"func": "Req()",
		}).Error("there are not current request")
	}
	return nil
}
