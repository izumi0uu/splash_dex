package chain

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/zeromicro/go-zero/core/logx"
)

func BlockFetcher(ctx context.Context, rpcUrl string, slotChan chan uint64) {
	// create Solana HTTP RPC client
	rpcClient := client.New(rpc.WithEndpoint(rpcUrl), rpc.WithHTTPClient(&http.Client{
		Timeout: 5 * time.Second,
	}))

	// read slot from slotChan
	for {
		select {
		case slot, ok := <-slotChan:
			if !ok {
				logx.Error("slotChan is closed")
				return
			}
			if slot == 0 {
				continue
			}
			processBlock(rpcClient, slot)
		case <-ctx.Done():
			logx.Info("BlockFetcher stopping...")
			return
		}
	}
}

func processBlock(rpcClient *client.Client, slot uint64) {
	// create a new context with timeout for every single request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	block, err := rpcClient.GetBlockWithConfig(ctx, slot, client.GetBlockConfig{
		Commitment:         rpc.CommitmentConfirmed,
		TransactionDetails: rpc.GetBlockConfigTransactionDetailsFull,
	})
	if err != nil {
		logx.Errorf("failed to get block: %v", err)
		return
	}
	fmt.Println("Block:", block)
}
