package svc

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"splash.xyz/dex/consumer/internal/config"
	"splash.xyz/dex/consumer/internal/model"
)

const defaultRPCTimeout = 30 * time.Second

type ServiceContext struct {
	Config         config.Config
	DB             sqlx.SqlConn
	SolBlockModel  model.SolBlocksModel
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
	return newServiceContext(c, nil)
}

func NewSolServiceContext(c config.Config) *ServiceContext {
	logx.MustSetup(c.Log)

	logx.Infof("newSolServiceContext: config: %v", c)

	// range all nodes, create client for each node
	var solClients []*rpc.Client
	for _, node := range c.Sol.NodeUrl {
		solClients = append(solClients, NewSolRPCClient(node))
	}

	return newServiceContext(c, solClients)
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

func newServiceContext(c config.Config, solClients []*rpc.Client) *ServiceContext {
	sc := &ServiceContext{
		Config:     c,
		solClients: solClients,
	}

	if conn := newMysqlConn(c.Mysql); conn != nil {
		sc.DB = conn
		sc.SolBlockModel = model.NewSolBlocksModel(conn)
	}

	return sc
}

func newMysqlConn(cfg config.Mysql) sqlx.SqlConn {
	if cfg.Host == "" || cfg.User == "" || cfg.Database == "" {
		logx.Infof("mysql config incomplete, block persistence disabled")
		return nil
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	return sqlx.NewMysql(dsn)
}
