package grpc

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/t1mon-ggg/go_shortner/app/grpc/proto"
	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
)

func stop(g *grpcServer) {
	g.app.Signal() <- syscall.SIGTERM
	time.Sleep(2 * time.Second)
	log.Println("terminated")
}

func Test_grpcServer_SimpleShort(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
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
}

func Test_grpcServer_APIStats(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
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
}

func Test_grpcServer_Ping(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.PingRequest)
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	_, err = app.Ping(ctx, request)
	require.NoError(t, err)
}

func Test_grpcServer_APIUserShortDelete(t *testing.T) {
	db := webhandlers.NewApp()
	err := db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.SimpleShortRequest)
	request.Url = "http://example.org"
	header := metadata.MD{}
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	short, err := app.SimpleShort(ctx, request, grpc.Header(&header))
	require.NoError(t, err)
	t.Log(header["client_id"])
	t.Log(short.Short)
	requestDel := new(pb.APIUserShortDeleteRequest)
	re := regexp.MustCompile(`\w{8}`)
	s := string(re.Find([]byte(short.Short)))
	requestDel.Urls = s
	ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Client_ID", header["client_id"][0]))
	_, err = app.APIUserShortDelete(ctx, requestDel)
	require.NoError(t, err)
	time.Sleep(15 * time.Second)
	requestdeleted := new(pb.SimpleUnshortRequest)
	requestdeleted.Short = s
	_, err = app.SimpleUnshort(ctx, requestdeleted)
	require.Error(t, err)
	t.Log(err)
}

func Test_grpcServer_APIUserURLs(t *testing.T) {
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}

	in := []input{
		{Correlation: "12345",
			Long: "http://example1.org"},
		{Correlation: "12345",
			Long: "http://example2.org"},
		{Correlation: "12345",
			Long: "http://example3.org"},
		{Correlation: "12345",
			Long: "http://example4.org"},
		{Correlation: "12345",
			Long: "http://example5.org"},
		{Correlation: "12345",
			Long: "http://example6.org"},
		{Correlation: "12345",
			Long: "http://example7.org"},
		{Correlation: "12345",
			Long: "http://example8.org"},
		{Correlation: "12345",
			Long: "http://example9.org"},
		{Correlation: "12345",
			Long: "http://example10.org"},
		{Correlation: "12345",
			Long: "http://example11.org"},
		{Correlation: "12345",
			Long: "http://example12.org"},
		{Correlation: "12345",
			Long: "http://example13.org"},
		{Correlation: "12345",
			Long: "http://example14.org"},
		{Correlation: "12345",
			Long: "http://example15.org"},
		{Correlation: "12345",
			Long: "http://example16.org"},
		{Correlation: "12345",
			Long: "http://example17.org"},
		{Correlation: "12345",
			Long: "http://example18.org"},
		{Correlation: "12345",
			Long: "http://example19.org"},
		{Correlation: "12345",
			Long: "http://example20.org"},
	}

	body, err := json.Marshal(in)
	require.NoError(t, err)

	db := webhandlers.NewApp()
	err = db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.APIBatchShortRequest)
	request.Jsonurls = string(body)
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	header := metadata.MD{}
	response, err := app.APIBatchShort(ctx, request, grpc.Header(&header))
	require.NoError(t, err)
	require.NotEmpty(t, response.Jsonshorts)
	ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Client_ID", header["client_id"][0]))
	urls := new(pb.APIUserURLRequest)
	response1, err := app.APIUserURLs(ctx, urls)
	require.NoError(t, err)
	require.NotEmpty(t, response1.Urls)
	t.Log(response1.Urls)

}

func Test_grpcServer_APIShort(t *testing.T) {
	type req struct {
		URL string `json:"url"` // {"url":"<some_url>"}
	}
	in := req{URL: "http://example.org"}
	body, err := json.Marshal(in)
	require.NoError(t, err)

	db := webhandlers.NewApp()
	err = db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.APIShortRequest)
	request.Jsonurl = string(body)
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	response, err := app.APIShort(ctx, request)
	require.NoError(t, err)
	require.NotEmpty(t, response.Jsonshort)
}

func Test_grpcServer_APIBatchShort(t *testing.T) {
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}

	in := []input{
		{Correlation: "12345",
			Long: "http://example1.org"},
		{Correlation: "12345",
			Long: "http://example2.org"},
		{Correlation: "12345",
			Long: "http://example3.org"},
		{Correlation: "12345",
			Long: "http://example4.org"},
		{Correlation: "12345",
			Long: "http://example5.org"},
		{Correlation: "12345",
			Long: "http://example6.org"},
		{Correlation: "12345",
			Long: "http://example7.org"},
		{Correlation: "12345",
			Long: "http://example8.org"},
		{Correlation: "12345",
			Long: "http://example9.org"},
		{Correlation: "12345",
			Long: "http://example10.org"},
		{Correlation: "12345",
			Long: "http://example11.org"},
		{Correlation: "12345",
			Long: "http://example12.org"},
		{Correlation: "12345",
			Long: "http://example13.org"},
		{Correlation: "12345",
			Long: "http://example14.org"},
		{Correlation: "12345",
			Long: "http://example15.org"},
		{Correlation: "12345",
			Long: "http://example16.org"},
		{Correlation: "12345",
			Long: "http://example17.org"},
		{Correlation: "12345",
			Long: "http://example18.org"},
		{Correlation: "12345",
			Long: "http://example19.org"},
		{Correlation: "12345",
			Long: "http://example20.org"},
	}

	body, err := json.Marshal(in)
	require.NoError(t, err)

	db := webhandlers.NewApp()
	err = db.NewStorage()
	require.NoError(t, err)
	g := New(db)
	go func() {
		go db.Start()
		g.Start()
	}()
	defer stop(g)
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app := pb.NewShortenerClient(conn)
	request := new(pb.APIBatchShortRequest)
	request.Jsonurls = string(body)
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	response, err := app.APIBatchShort(ctx, request)
	require.NoError(t, err)
	require.NotEmpty(t, response.Jsonshorts)
	t.Log(response.Jsonshorts)
}
