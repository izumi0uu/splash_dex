package config

import "github.com/zeromicro/go-zero/zrpc"

var Cfg Config

var (
	SolRpcUseFrequency int
)

type Config struct {
	zrpc.RpcServerConf

	Sol Chain `json:"Sol,optional"`
}

type Chain struct {
	ChainId    int64    `json:"ChainId"`
	NodeUrl    []string `json:"NodeUrl"`             // http rpc node lists, support multiple node urls
	MEVNodeUrl string   `json:"MevNodeUrl,optional"` // MEV protected node url
	WSUrl      string   `json:"WSUrl,optional"`      // websocket url
	StartBlock uint64   `json:"StartBlock,optional"` // start block number
}
