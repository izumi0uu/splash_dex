package block

import (
	"encoding/binary"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/program/token"

	"splash.xyz/dex/pkg/constants"
)

// Block
//  └── Transaction
//       ├── Instructions→ recongnize which DEX program
//       │    └── InnerInstructions（CPI）→ DEX program call SPL Token
//       │         └── Instruction Type（Transfer / TransferChecked）→ parse amount
//       └── Meta
//            ├── PreTokenBalances  → token account → mint mapping
//            └── PostTokenBalances → supplement temporary accounts

func findInnerInstructions(tx *client.BlockTransaction, ixIndex int) *client.InnerInstruction {
	for i := range tx.Meta.InnerInstructions {
		if tx.Meta.InnerInstructions[i].Index == uint64(ixIndex) {
			return &tx.Meta.InnerInstructions[i]
		}
	}
	return nil
}

// extract all token transfers from inner instructions
func extractTokenTransfers(tx *client.BlockTransaction, innerIx *client.InnerInstruction) []*token.TransferParam {
	var transfers []*token.TransferParam

	for _, ix := range innerIx.Instructions {
		// is array out of bounds?
		if int(ix.ProgramIDIndex) >= len(tx.AccountKeys) {
			continue
		}
		program := tx.AccountKeys[ix.ProgramIDIndex].String()

		// only address SPL Token's Transfer instruction
		if program != constants.ProgramStrToken && program != constants.ProgramStrToken2022 {
			continue
		}
		// is empty?
		if len(ix.Data) < 1 {
			continue
		}

		switch token.Instruction(ix.Data[0]) {
		// SPL Token Transfer (discriminator = 3)
		//   Data:     [type: 1B] [amount: 8B LE]         → min 9 bytes
		//   Accounts: [0]=from  [1]=to  [2]=authority    → min 3 accounts
		case token.InstructionTransfer:
			if len(ix.Data) < 9 || len(ix.Accounts) < 3 {
				continue
			}
			transfers = append(transfers, &token.TransferParam{
				From:   tx.AccountKeys[ix.Accounts[0]],
				To:     tx.AccountKeys[ix.Accounts[1]],
				Auth:   tx.AccountKeys[ix.Accounts[2]],
				Amount: binary.LittleEndian.Uint64(ix.Data[1:9]),
			})
		// SPL Token TransferChecked (discriminator = 12)
		//   Data:     [type: 1B] [amount: 8B LE] [decimals: 1B]  → min 10 bytes
		//   Accounts: [0]=from  [1]=mint  [2]=to  [3]=authority  → min 4 accounts
		case token.InstructionTransferChecked:
			if len(ix.Data) < 10 || len(ix.Accounts) < 4 {
				continue
			}
			transfers = append(transfers, &token.TransferParam{
				From: tx.AccountKeys[ix.Accounts[0]],
				// Accounts[1] = mint, skipped
				To:     tx.AccountKeys[ix.Accounts[2]],
				Auth:   tx.AccountKeys[ix.Accounts[3]],
				Amount: binary.LittleEndian.Uint64(ix.Data[1:9]),
			})
		}
	}
	// fill token mints(Transfer instruction doesn't contain mint, need to query from PreTokenBalances)

	return transfers
}
