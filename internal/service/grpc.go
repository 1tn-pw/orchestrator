package service

import (
	"github.com/1tn-pw/orchestrator/internal/config"
	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/bugfixes/go-bugfixes/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ShortService interface {
	NewShortService() pb.ShortServiceClient
}

type Short struct {
	config.Config
	context.Context
	Client pb.ShortServiceClient
}

func NewShortService(ctx context.Context, cfg *config.Config) *Short {
	return &Short{
		Config:  *cfg,
		Context: ctx,
	}
}

func (s *Short) GetLong(ctx context.Context, short string) (*pb.GetURLResponse, error) {
	conn, err := grpc.DialContext(ctx, s.Config.Services.ShortService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &pb.GetURLResponse{
			Error: utils.Pointer(err.Error()),
		}, nil
	}

	defer func() {
		if err := conn.Close(); err != nil {
			_ = logs.Errorf("Error closing connection: %v", err)
		}
	}()

	client := pb.NewShortServiceClient(conn)
	return client.GetURL(ctx, &pb.GetURLRequest{
		ShortUrl: short,
	})
}

func (s *Short) CreateShort(ctx context.Context, long string) (*pb.CreateURLResponse, error) {
	conn, err := grpc.DialContext(ctx, s.Config.Services.ShortService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &pb.CreateURLResponse{
			Error: utils.Pointer(err.Error()),
		}, nil
	}

	defer func() {
		if err := conn.Close(); err != nil {
			_ = logs.Errorf("Error closing connection: %v", err)
		}
	}()

	client := pb.NewShortServiceClient(conn)
	return client.CreateURL(ctx, &pb.CreateURLRequest{
		Url: long,
	})
}
