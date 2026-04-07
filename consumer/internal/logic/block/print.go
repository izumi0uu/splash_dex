package block

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func PrintPumpFunBondingCurve(txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	fmt.Printf("[PumpFun BC] tx: %s, ixIndex: %d\n", txHash, ixIndex)
}

func PrintPumpSwapAMM(txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	fmt.Printf("[PumpSwap] tx: %s, ixIndex: %d\n", txHash, ixIndex)
}
