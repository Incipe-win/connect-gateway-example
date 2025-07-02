package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"

	proto "bookstore_enhance/proto/gateway"
	"bookstore_enhance/proto/protoconnect"
	"bookstore_enhance/third_party"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	serviceName string = "bookstore"
	port        int    = 8888
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		panic("couldn't create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFS))
}

func run() error {
	log := grpclog.NewLoggerV2(os.Stdout, io.Discard, io.Discard)
	grpclog.SetLoggerV2(log)

	// 初始化数据库
	db, err := NewDB("bookstore.db")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	addr := fmt.Sprintf(":%d", port)
	conn, err := grpc.NewClient(
		"dns:///"+addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
			},
		}),
	)
	err = proto.RegisterBookStoreHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gRPC gateway: %w", err)
	}

	// 连接到 consul
	cc, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create Consul client: %w", err)
	}
	localIP, err := GetOutboundIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP: %w", err)
	}
	// 定义我们的服务
	srv := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", serviceName, localIP.String(), port),
		Name:    serviceName,
		Tags:    []string{"hchao"},
		Address: localIP.String(),
		Port:    port,
	}
	// 注册服务到 Consul
	cc.Agent().ServiceRegister(srv)

	mux := http.NewServeMux()
	mux.Handle("/", getOpenAPIHandler())
	mux.Handle(protoconnect.NewBookStoreHandler(&server{bs: &bookstore{db: db}}))
	mux.Handle("/api/v1/", gwmux)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	log.Info("Starting HTTP server on ", addr)
	return server.ListenAndServe()
}

// GetOutboundIP 获取本机的出口IP
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
