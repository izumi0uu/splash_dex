package svc

import (
	"context"
	"splash.xyz/dex/consumer/internal/chain"
	"splash.xyz/dex/consumer/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	ChainClient *chain.Client
	SlotChan    chan uint64
	Cancel      context.CancelFunc
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx, cancel := context.WithCancel(context.Background())

	client := chain.NewClient(c.Sol.WSUrl)
	slotChan := make(chan uint64, 100)

	go chain.SlotListener(client, slotChan)
	go chain.BlockFetcher(ctx, c.Sol.NodeUrl[0], slotChan)

	return &ServiceContext{
		Config:      c,
		ChainClient: client,
		SlotChan:    slotChan,
		Cancel:      cancel,
	}
}
