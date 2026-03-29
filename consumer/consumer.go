package main

import (
	"flag"
	"fmt"

	"splash.xyz/dex/consumer/consumer"
	"splash.xyz/dex/consumer/internal/config"
	"splash.xyz/dex/consumer/internal/logic/block"
	"splash.xyz/dex/consumer/internal/logic/slot"
	"splash.xyz/dex/consumer/internal/server"
	"splash.xyz/dex/consumer/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/consumer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	config.Cfg = c // sync to global so FindChainRpcByChainId can read it
	ctx := svc.NewSolServiceContext(c)

	// manage multiple services
	group := service.NewServiceGroup()
	defer group.Stop()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		consumer.RegisterConsumerServer(grpcServer, server.NewConsumerServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	group.Add(s)

	{
		// increment message queue
		slotChan := make(chan uint64, 100)

		// consumer: comsuming concurrently
		for i := 0; i < c.Consumer.Concurrency; i++ {
			group.Add(block.NewBlockService(ctx, "block-real", slotChan, i))
		}

		// Producer: get latest slot
		group.Add(slot.NewSlotServiceGroup(ctx, slotChan))
	}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}
