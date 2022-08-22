package grpc

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/t1mon-ggg/go_shortner/app/grpc/proto"
	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
)

func Test_grpcServer_SimpleShort(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	go func(db *webhandlers.App) {
		err := StartGRPC(db)
		require.NoError(t, err)
	}(db)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.SimpleShortRequest)
	request.Url = "http://example.org"
	var header metadata.MD
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	response, err := app.SimpleShort(ctx, request, grpc.Trailer(&header))
	require.NoError(t, err)
	require.NotEmpty(t, response.Short)
	t.Log(header)
	t.Log(response.Short)
}

func Test_grpcServer_APIStats(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	go func(db *webhandlers.App) {
		err := StartGRPC(db)
		require.NoError(t, err)
	}(db)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.APIStatsRequest)
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	_, err = app.APIStats(ctx, request)
	require.Error(t, err)
	t.Log(err)
}
