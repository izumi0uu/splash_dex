package config

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	HeliusWebsocketUrl string `yaml:"helius_websocket_url" json:"helius_websocket_url" env:"HELIUS_WEBSOCKET_URL"`
	HeliusHttpRpcUrl   string `yaml:"helius_http_rpc_url" json:"helius_http_rpc_url" env:"HELIUS_HTTP_RPC_URL"`
}

var Cfg Config

func LoadConfig(configFile string) error {
	return conf.Load(configFile, &Cfg)
}
