package block

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// handleDexSwap is the generic handler for DEX programs without Anchor events (e.g. Raydium V4).
// It extracts token transfers from inner instructions and identifies the tokens involved.
// For Anchor-based programs (PumpFun, PumpSwap, CLMM, CPMM), use protocol-specific handlers instead.
func (s *BlockService) handleDexSwap(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int, dexName string) {
	// Step 1: find inner instructions (CPI calls)
	innerIxs := findInnerInstructions(meta, ixIndex)
	if innerIxs == nil {
		return
	}

	// Step 2: extract Token Transfer from inner instructions
	transfers := extractTokenTransfers(accountKeys, innerIxs)
	if len(transfers) < 2 {
		return // a swap requires at least 2 transfers (one in, one out)
	}

	// Step 3: build mapping for token mint (token account → mint address)
	mintMap := buildMintMap(meta, accountKeys)

	// Step 4: print (replace with actual business logic later)
	fmt.Printf("[%s] tx: %s, transfers: %d\n", dexName, txHash, len(transfers))
	for _, t := range transfers {
		tokenMint := mintMap[t.From.String()]
		fmt.Printf("  %s → %s, amount: %d, token: %s\n", t.From, t.To, t.Amount, tokenMint)
	}
}
