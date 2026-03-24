package config

import (
	"splash.xyz/dex/pkg/constants"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var Cfg Config

var (
	SolRpcUseFrequency int
)

type Config struct {
	zrpc.RpcServerConf

	Sol Chain `json:"Sol,optional"`

	Consumer Consumer `json:"Consumer,optional"`
}

type Consumer struct {
	Concurrency int `json:"Concurrency" json:", env=CONSUMER_CONCURRENCY"`
}

type Chain struct {
	ChainId    int64    `json:"ChainId"`
	NodeUrl    []string `json:"NodeUrl"`             // http rpc node lists, support multiple node urls
	MEVNodeUrl string   `json:"MevNodeUrl,optional"` // MEV protected node url
	WSUrl      string   `json:"WSUrl,optional"`      // websocket url
	StartBlock uint64   `json:"StartBlock,optional"` // start block number
}

func FindChainRpcByChainId(chainId int) (rpc string) {
	var rpcNodes []string
	var useFrequency *int

	switch chainId {
	case constants.SolChainIdInt:
		rpcNodes = Cfg.Sol.NodeUrl
		useFrequency = &SolRpcUseFrequency
	default:
		logx.Errorf("No RPC Config for chainId: %d", chainId)
		return
	}

	if len(rpcNodes) == 0 {
		logx.Errorf("No RPC Config for chainId: %d", chainId)
		return
	}

	// polling to choose rpc node(Load Balancing Core)
	*useFrequency++
	index := *useFrequency % len(rpcNodes) // get Modulo
	rpc = rpcNodes[index]
	return
}
