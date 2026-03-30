package svc

import (
	"net/http"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"splash.xyz/dex/consumer/internal/config"
)

const defaultRPCTimeout = 10 * time.Second

type ServiceContext struct {
	Config         config.Config
	solClientLock  sync.Mutex
	solClientIndex int
	solClient      *rpc.Client
	solClients     []*rpc.Client
}

func NewSolRPCClient(endpoint string) *rpc.Client {
	rpcClient := jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
		HTTPClient: &http.Client{
			Timeout: defaultRPCTimeout,
		},
	})
	return rpc.NewWithCustomRPCClient(rpcClient)
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func NewSolServiceContext(c config.Config) *ServiceContext {
	logx.MustSetup(c.Log)

	logx.Infof("newSolServiceContext: config: %v", c)

	// range all nodes, create client for each node
	var solClients []*rpc.Client
	for _, node := range c.Sol.NodeUrl {
		solClients = append(solClients, NewSolRPCClient(node))
	}

	return &ServiceContext{
		Config:     c,
		solClients: solClients,
	}
}

func (sc *ServiceContext) GetSolClient() *rpc.Client {
	// 1. lock to prevent managing goroutines concurrently
	sc.solClientLock.Lock()
	defer sc.solClientLock.Unlock()

	sc.solClientIndex++

	// 2. get modulo
	index := sc.solClientIndex % len(sc.solClients)
	sc.solClient = sc.solClients[index]
	return sc.solClients[index]
}
