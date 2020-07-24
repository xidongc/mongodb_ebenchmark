package main

import (
	"context"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/xidongc/mongodb_ebenchmark/model/order/orderpb"
	order "github.com/xidongc/mongodb_ebenchmark/model/order/service"
	"github.com/xidongc/mongodb_ebenchmark/model/payment/paymentpb"
	payment "github.com/xidongc/mongodb_ebenchmark/model/payment/service"
	"github.com/xidongc/mongodb_ebenchmark/model/product/productpb"
	product "github.com/xidongc/mongodb_ebenchmark/model/product/service"
	sku "github.com/xidongc/mongodb_ebenchmark/model/sku/service"
	"github.com/xidongc/mongodb_ebenchmark/model/sku/skupb"
	user "github.com/xidongc/mongodb_ebenchmark/model/user/service"
	"github.com/xidongc/mongodb_ebenchmark/model/user/userpb"
	"github.com/xidongc/mongodb_ebenchmark/pkg/proxy"
	server "github.com/xidongc/mongodb_ebenchmark/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var config server.Config

	parser := flags.NewParser(&config, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
	log.Infof("%+v", config)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	maxSendMsgSize := 1024 * 1024 * 500
	maxRecvMsgSize := 1024 * 1024 * 100

	maxSendMsgSizeOpt := grpc.MaxSendMsgSize(maxSendMsgSize)
	maxRecvMsgSizeOpt := grpc.MaxRecvMsgSize(maxRecvMsgSize)

	svr := grpc.NewServer(maxSendMsgSizeOpt, maxRecvMsgSizeOpt)
	proxyConfig := &proxy.Config{
		ServerIp:     config.ServerAddr,
		Port:         config.ServerPort,
		Insecure:     config.Insecure,
		RpcTimeout:   config.RpcTimeout,
		BatchSize:    config.BatchSize,
		ReadPref:     config.ReadPref,
		AllowPartial: config.AllowPartial,
	}
	storageClient := *sku.NewClient(proxyConfig, cancel)

	if config.Turbo {
		storageClient.Turbo = true
	}

	defer func() {
		err := storageClient.Close()
		if err != nil {
			log.Error("error")
		}
	}()


	amplifyOptions := &proxy.AmplifyOptions{
		Connections:  config.Connections,
		Concurrency:  config.Concurrency,
		TotalRequest: config.TotalRequest,
		QPS:          config.QPS,
		Timeout:      config.Timeout,
		CPUs:         config.CPUs,
	}

	skuService := &sku.Service{
		Storage: storageClient,
		Amplifier: amplifyOptions,
	}
	paymentService := &payment.Service{
		Storage: storageClient,
		Amplifier: amplifyOptions,
	}

	orderService := &order.Service{
		Storage: storageClient,
		Amplifier: amplifyOptions,
	}

	userService := &user.Service {
		Storage: storageClient,
		Amplifier: amplifyOptions,
	}

	productService := &product.Service {
		Storage: storageClient,
		Amplifier: amplifyOptions,
	}

	skupb.RegisterSkuServiceServer(svr, skuService)
	paymentpb.RegisterPaymentServiceServer(svr, paymentService)
	orderpb.RegisterOrderServiceServer(svr, orderService)
	productpb.RegisterProductServiceServer(svr, productService)
	userpb.RegisterUserServiceServer(svr, userService)

	reflection.Register(svr)

	go func() {
		addr := fmt.Sprintf("%s:%d", config.ServerAddr, config.ServerPort)
		log.Infof("Start listening on %s", addr)
		lis, err := net.Listen("tcp4", addr)
		if err != nil {
			log.Fatal(err)
		}
		if err = svr.Serve(lis); err != nil {
			log.Fatal(err)
		}
		cancel()
	}()
	select {
	case <-sigs:
	case <-ctx.Done():
	}

	log.Warn("Got shutdown signal")
	svr.GracefulStop()
}