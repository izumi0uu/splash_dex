package block

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	bonding_curve "splash.xyz/dex/pkg/contracts/pumpfun/bondingcurve/idl/generated"
	amm "splash.xyz/dex/pkg/contracts/pumpfun/amm/idl/generated"
)

// PumpFun Bonding Curve token decimals (always 6 for tokens created via pump.fun)
const pumpFunTokenDecimals = 6

// handlePumpFunBondingCurve handles swap instructions from the PumpFun Bonding Curve program.
// Uses Anchor Event parsing (TradeEvent) from LogMessages.
func (s *BlockService) handlePumpFunBondingCurve(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	events := extractAnchorEvents(meta.LogMessages)
	if len(events) == 0 {
		return
	}

	for _, eventData := range events {
		tradeEvent, err := bonding_curve.ParseEvent_TradeEvent(eventData)
		if err != nil {
			continue // not a TradeEvent (could be CreateEvent, CompleteEvent, etc.)
		}

		direction := "sell"
		if tradeEvent.IsBuy {
			direction = "buy"
		}

		// TODO: replace with actual business logic (build TradeWithPair, save to DB, etc.)
		fmt.Printf("[PumpFun BC] %s tx: %s, mint: %s, sol: %d, token: %d, user: %s, fee: %d\n",
			direction, txHash,
			tradeEvent.Mint,
			tradeEvent.SolAmount,
			tradeEvent.TokenAmount,
			tradeEvent.User,
			tradeEvent.Fee,
		)
	}
}

// handlePumpSwapAMM handles swap instructions from the PumpSwap AMM program.
// Uses Anchor Event parsing (BuyEvent / SellEvent) from LogMessages.
func (s *BlockService) handlePumpSwapAMM(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	events := extractAnchorEvents(meta.LogMessages)
	if len(events) == 0 {
		return
	}

	for _, eventData := range events {
		parsed, err := amm.ParseAnyEvent(eventData)
		if err != nil {
			continue
		}

		switch evt := parsed.(type) {
		case *amm.BuyEvent:
			fmt.Printf("[PumpSwap] buy tx: %s, pool: %s, user: %s, baseOut: %d, quoteIn: %d, lpFee: %d, protocolFee: %d\n",
				txHash,
				evt.Pool,
				evt.User,
				evt.BaseAmountOut,
				evt.QuoteAmountIn,
				evt.LpFee,
				evt.ProtocolFee,
			)
		case *amm.SellEvent:
			fmt.Printf("[PumpSwap] sell tx: %s, pool: %s, user: %s, baseIn: %d, quoteOut: %d, lpFee: %d, protocolFee: %d\n",
				txHash,
				evt.Pool,
				evt.User,
				evt.BaseAmountIn,
				evt.QuoteAmountOut,
				evt.LpFee,
				evt.ProtocolFee,
			)
		case *amm.CreatePoolEvent:
			fmt.Printf("[PumpSwap] createPool tx: %s, pool: %s, creator: %s\n",
				txHash,
				evt.Pool,
				evt.Creator,
			)
		}
	}
}
