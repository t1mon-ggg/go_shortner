package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"

	pb "github.com/t1mon-ggg/go_shortner/app/grpc/proto"
	"github.com/t1mon-ggg/go_shortner/app/helpers"
	"github.com/t1mon-ggg/go_shortner/app/models"
	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func tokenGen(key, value string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	signed := h.Sum(nil)
	return hex.EncodeToString(signed)
}

func saveToken(ctx context.Context) string {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("Client_ID")
		if len(values) > 0 {
			token = values[0]
		}
	}
	return token
}

func StartGRPC(app *webhandlers.App) error {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return err
	}
	grpcEndpoint := new(grpcServer)
	grpcEndpoint.app = app
	s := grpc.NewServer(grpc.UnaryInterceptor(grpcEndpoint.grpcCookie))
	pb.RegisterShortenerServer(s, grpcEndpoint)
	log.Println("GRPC server started")
	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}

type grpcServer struct {
	app *webhandlers.App
	pb.UnimplementedShortenerServer
}

func (server *grpcServer) checkToken(token string) bool {
	data := token[:32]
	signstring := token[32:]
	sign, err := hex.DecodeString(signstring)
	if err != nil {
		log.Println(err)
		return false
	}
	checkdata, _ := server.app.Storage.ReadByCookie(data)
	h := hmac.New(sha256.New, []byte(checkdata.Key))
	h.Write([]byte(data))
	signed := h.Sum(nil)
	return hmac.Equal(sign, signed)
}

func (server *grpcServer) grpcCookie(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("Client_ID")
		if len(values) > 0 {
			log.Println("Client_ID found. Checking...")
			token = values[0]
			if !server.checkToken(token) {
				log.Println("Token check failed. Regenerating")
				value := helpers.RandStringRunes(32)
				key := helpers.RandStringRunes(64)
				log.Println("New token:", tokenGen(key, value))
				newmd := metadata.New(map[string]string{"Client_ID": tokenGen(key, value)})
				newctx := metadata.NewIncomingContext(context.Background(), newmd)
				return handler(newctx, req)
			}

		} else {
			log.Println("Client_ID not found. Generating")
			value := helpers.RandStringRunes(32)
			key := helpers.RandStringRunes(64)
			log.Println("New token:", tokenGen(key, value))
			newmd := metadata.Pairs("Client_ID", tokenGen(key, value))
			newctx := metadata.NewIncomingContext(ctx, newmd)
			return handler(newctx, req)
		}
	}
	return handler(ctx, req)
}

func (server *grpcServer) SimpleShort(ctx context.Context, in *pb.SimpleShortRequest) (*pb.SimpleShortResponse, error) {
	cookie := saveToken(ctx)
	defer func() {
		trailer := metadata.Pairs("Client_ID", cookie)
		err := grpc.SetTrailer(ctx, trailer)
		if err != nil {
			log.Println("trailer error:", err)
		}
	}()
	entry := models.ClientData{}
	entry.Cookie = cookie
	entry.Key = ""
	entry.Short = make([]models.ShortData, 0)
	long := in.Url
	short := helpers.RandStringRunes(8)
	entry.Short = append(entry.Short, models.ShortData{Short: short, Long: long})
	log.Println(entry)
	err := server.app.Storage.Write(entry)
	if err != nil {
		if err.Error() == "not unique url" {
			s, errTag := server.app.Storage.TagByURL(long, cookie)
			if errTag != nil {
				log.Println(errTag)
				return nil, status.Error(codes.Unknown, "storage error")
			}
			return &pb.SimpleShortResponse{Short: fmt.Sprintf("%s/%s", server.app.Config.BaseURL, s)}, status.Error(codes.Internal, "Already exists")
		}
	}
	header := metadata.Pairs("Client_ID", cookie)
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		log.Println("header error:", err)
	}
	response := new(pb.SimpleShortResponse)
	response.Short = fmt.Sprintf("%s/%s", server.app.Config.BaseURL, short)
	return response, nil
}

func (server *grpcServer) SimpleUnshort(ctx context.Context, in *pb.SimpleUnshortRequest) (*pb.SimpleUnshortResponse, error) {
	response := new(pb.SimpleUnshortResponse)
	cookie := saveToken(ctx)
	defer func() {
		trailer := metadata.Pairs("Client_ID", cookie)
		err := grpc.SetTrailer(ctx, trailer)
		if err != nil {
			log.Println("trailer error:", err)
		}
	}()
	matched, err := regexp.Match(`\w{8}`, []byte(in.Short))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid short uri")
	}
	re := regexp.MustCompile(`\w{8}`)
	r := re.FindAllString(in.Short, 1)
	if len(r) != 1 {
		return nil, status.Error(codes.InvalidArgument, "invalid short uri")
	}
	p := r[0]
	if !matched || len(p) > 8 {
		return nil, status.Error(codes.InvalidArgument, "invalid short uri")
	}
	data, err := server.app.Storage.ReadByTag(p)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB read error")
	}
	nilShort := models.ShortData{}
	if data == nilShort {
		return nil, status.Error(codes.NotFound, "NotFound")
	}
	if data.Deleted {
		return nil, status.Error(codes.OutOfRange, "Deleted")
	}
	header := metadata.Pairs("Client_ID", cookie)
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		log.Println("header error:", err)
	}
	response.Url = data.Long
	return response, nil
}

func (server *grpcServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	err := server.app.Storage.Ping()
	if err != nil {
		return nil, status.Error(codes.DataLoss, err.Error())
	}
	return nil, nil
}

func (server *grpcServer) APIUserURLs(ctx context.Context, in *pb.APIUserURLRequest) (*pb.APIUserURLResponse, error) {
	response := new(pb.APIUserURLResponse)
	cookie := saveToken(ctx)
	defer func() {
		trailer := metadata.Pairs("Client_ID", cookie)
		err := grpc.SetTrailer(ctx, trailer)
		if err != nil {
			log.Println("trailer error:", err)
		}
	}()
	header := metadata.Pairs("Client_ID", cookie)
	err := grpc.SendHeader(ctx, header)
	if err != nil {
		log.Println("header error:", err)
	}
	data, err := server.app.Storage.ReadByCookie(cookie)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(data.Short) == 0 {
		return nil, status.Error(codes.NotFound, "No content")
	}
	type answer struct {
		Short    string `json:"short_url"`
		Original string `json:"original_url"`
	}
	a := make([]answer, 0)
	for _, content := range data.Short {
		a = append(a, answer{Short: fmt.Sprintf("%s/%s", server.app.Config.BaseURL, content.Short), Original: content.Long})
	}
	d, err := json.Marshal(a)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(a) == 0 {
		return nil, status.Error(codes.NotFound, "No content")
	}
	response.Urls = string(d)
	return response, nil
}
func (server *grpcServer) APIStats(ctx context.Context, in *pb.APIStatsRequest) (*pb.APIStatsResponse, error) {
	if server.app.Config.TrustedSubnet == "" {
		log.Println("endpoint locked")
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}
	response := new(pb.APIStatsResponse)
	p, _ := peer.FromContext(ctx)
	ip := p.Addr.String()
	_, tSubnet, err := net.ParseCIDR(server.app.Config.TrustedSubnet)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unknown, "CIDR parse error")
	}
	if !tSubnet.Contains(net.ParseIP(ip)) {
		log.Println("ip not in trust subnet")
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}
	stats, err := server.app.Storage.GetStats()
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unknown, "get stats internal storage error")
	}
	sts, err := json.Marshal(stats)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unknown, "json marshal stats error")
	}
	response.Stats = string(sts)
	return response, nil
}
func (server *grpcServer) APIShort(ctx context.Context, in *pb.APIShortRequest) (*pb.APIShortResponse, error) {
	response := new(pb.APIShortResponse)
	cookie := saveToken(ctx)
	defer func() {
		trailer := metadata.Pairs("Client_ID", cookie)
		err := grpc.SetTrailer(ctx, trailer)
		if err != nil {
			log.Println("trailer error:", err)
		}
	}()
	header := metadata.Pairs("Client_ID", cookie)
	err := grpc.SendHeader(ctx, header)
	if err != nil {
		log.Println("header error:", err)
	}
	entry := models.ClientData{}
	entry.Cookie = cookie
	entry.Key = ""
	entry.Short = make([]models.ShortData, 0)
	type lURL struct {
		LongURL string `json:"url"`
	}
	type sURL struct {
		ShortURL string `json:"result"`
	}
	long := lURL{}
	err = json.Unmarshal([]byte(in.Jsonurl), &long)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		return nil, status.Error(codes.Internal, "JSON unmarshal error")
	}
	short := helpers.RandStringRunes(8)
	entry.Short = append(entry.Short, models.ShortData{Short: short, Long: long.LongURL})
	answer := sURL{}
	err = server.app.Storage.Write(entry)
	if err != nil {
		if err.Error() == "not unique url" {
			s, errTag := server.app.Storage.TagByURL(long.LongURL, cookie)
			if errTag != nil {
				log.Println(errTag)
				return nil, status.Error(codes.Unknown, "storage error")
			}
			answer.ShortURL = fmt.Sprintf("%s/%s", server.app.Config.BaseURL, s)
			j, err := json.Marshal(answer)
			if err != nil {
				log.Println("JSON Unmarshal error", err)
				return nil, status.Error(codes.Internal, "JSON marshal error")
			}
			return &pb.APIShortResponse{Jsonshort: string(j)}, status.Error(codes.Internal, "Already exists")
		}
	}
	answer.ShortURL = fmt.Sprintf("%s/%s", server.app.Config.BaseURL, short)
	j, err := json.Marshal(answer)
	if err != nil {
		log.Println("JSON marshal error", err)
		return nil, status.Error(codes.Internal, "JSON marshal error")
	}
	response.Jsonshort = string(j)
	return response, nil
}

func (server *grpcServer) APIBatchShort(ctx context.Context, in *pb.APIBatchShortRequest) (*pb.APIBatchShortResponse, error) {
	response := new(pb.APIBatchShortResponse)
	cookie := saveToken(ctx)
	defer func() {
		trailer := metadata.Pairs("Client_ID", cookie)
		err := grpc.SetTrailer(ctx, trailer)
		if err != nil {
			log.Println("trailer error:", err)
		}
	}()
	header := metadata.Pairs("Client_ID", cookie)
	err := grpc.SendHeader(ctx, header)
	if err != nil {
		log.Println("header error:", err)
	}
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}
	out := make([]output, 0)
	jsonin := []input{}
	err = json.Unmarshal([]byte(in.Jsonurls), &jsonin)
	if err != nil {
		log.Println("JSON unmarshal error", err)
		return nil, status.Error(codes.Internal, "JSON unmarshal error")
	}
	for i := range jsonin {
		entry := models.ClientData{}
		entry.Cookie = cookie
		entry.Key = ""
		entry.Short = make([]models.ShortData, 0)
		short := helpers.RandStringRunes(8)
		entry.Short = append(entry.Short, models.ShortData{Short: short, Long: jsonin[i].Long})
		errWrite := server.app.Storage.Write(entry)
		if errWrite != nil {
			if errWrite.Error() == "not unique url" {
				s, errTag := server.app.Storage.TagByURL(jsonin[i].Long, cookie)
				if errTag != nil {
					log.Println("storage error", err)
					return nil, status.Error(codes.Internal, "storage error")
				}
				out = append(out, output{Correlation: jsonin[i].Correlation, Short: fmt.Sprintf("%s/%s", server.app.Config.BaseURL, s)})
			} else {
				log.Println("storage error", err)
				return nil, status.Error(codes.Internal, "storage error")
			}
		} else {
			out = append(out, output{Correlation: jsonin[i].Correlation, Short: fmt.Sprintf("%s/%s", server.app.Config.BaseURL, short)})
		}
	}
	batch, err := json.Marshal(out)
	if err != nil {
		log.Println("json error", err)
		return nil, status.Error(codes.Internal, "json error")
	}
	response.Jsonshorts = string(batch)
	return response, nil
}
func (server *grpcServer) APIUserShortDelete(ctx context.Context, in *pb.APIUserShortDeleteRequest) (*pb.APIUserShortDeleteResponse, error) {
	response := new(pb.APIUserShortDeleteResponse)
	cookie := saveToken(ctx)
	re := regexp.MustCompile(`\w+`)
	tags := re.FindAllString(in.Urls, -1)
	task := models.DelWorker{Cookie: cookie, Tags: tags}
	server.app.DelBuf <- task
	return response, nil
}
