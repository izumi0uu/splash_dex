package block

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

func GetSolBlockInfoDelay(c *rpc.Client, ctx context.Context, slot uint64) (resp *rpc.GetBlockResult, err error) {
	// to reduce helius call
	time.Sleep(time.Second * 1)
	return GetSolBlockInfo(c, ctx, slot)
}

func GetSolBlockInfo(c *rpc.Client, ctx context.Context, slot uint64) (resp *rpc.GetBlockResult, err error) {
	var count int64
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// get block info
		resp, err = c.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			TransactionDetails:             rpc.TransactionDetailsFull,
			MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion0,
		})

		switch {
		case err == nil:
			return
		case strings.Contains(err.Error(), "Block not available for slot"):
			count++
			if count > 10 {
				return nil, err
			}
			time.Sleep(time.Second)
		case strings.Contains(err.Error(), "limit"):
			count++
			if count > 10 {
				return nil, err
			}
			time.Sleep(time.Second)
		case strings.Contains(err.Error(), "not confirmed"):
			count++
			if count > 10 {
				return nil, err
			}
			time.Sleep(time.Second)
		default:
			err = fmt.Errorf("get block info error: %v", err)
			return nil, err
		}
	}
}
