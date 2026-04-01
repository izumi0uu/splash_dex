package block

import (
	"encoding/base64"
	"encoding/binary"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

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

func findInnerInstructions(meta *rpc.TransactionMeta, ixIndex int) *rpc.InnerInstruction {
	for i := range meta.InnerInstructions {
		if meta.InnerInstructions[i].Index == uint16(ixIndex) {
			return &meta.InnerInstructions[i]
		}
	}
	return nil
}

// extract all token transfers from inner instructions
func extractTokenTransfers(accountKeys solana.PublicKeySlice, innerIx *rpc.InnerInstruction) []*TokenTransfer {
	var transfers []*TokenTransfer

	for _, ix := range innerIx.Instructions {
		// is array out of bounds?
		if int(ix.ProgramIDIndex) >= len(accountKeys) {
			continue
		}
		program := accountKeys[ix.ProgramIDIndex].String()

		// only address SPL Token's Transfer instruction
		if program != constants.ProgramStrToken && program != constants.ProgramStrToken2022 {
			continue
		}
		// is empty?
		if len(ix.Data) < 1 {
			continue
		}

		switch ix.Data[0] {
		// SPL Token Transfer (discriminator = 3)
		//   Data:     [type: 1B] [amount: 8B LE]         → min 9 bytes
		//   Accounts: [0]=from  [1]=to  [2]=authority    → min 3 accounts
		case constants.InstructionTransfer:
			if len(ix.Data) < 9 || len(ix.Accounts) < 3 {
				continue
			}
			transfers = append(transfers, &TokenTransfer{
				From:   accountKeys[ix.Accounts[0]],
				To:     accountKeys[ix.Accounts[1]],
				Auth:   accountKeys[ix.Accounts[2]],
				Amount: binary.LittleEndian.Uint64(ix.Data[1:9]),
			})
		// SPL Token TransferChecked (discriminator = 12)
		//   Data:     [type: 1B] [amount: 8B LE] [decimals: 1B]  → min 10 bytes
		//   Accounts: [0]=from  [1]=mint  [2]=to  [3]=authority  → min 4 accounts
		case constants.InstructionTransferChecked:
			if len(ix.Data) < 10 || len(ix.Accounts) < 4 {
				continue
			}
			transfers = append(transfers, &TokenTransfer{
				From: accountKeys[ix.Accounts[0]],
				// Accounts[1] = mint, skipped
				To:     accountKeys[ix.Accounts[2]],
				Auth:   accountKeys[ix.Accounts[3]],
				Amount: binary.LittleEndian.Uint64(ix.Data[1:9]),
			})
		}
	}

	return transfers
}

const logPrefixProgramData = "Program data: "

// extractAnchorEvents collects all Anchor event payloads from log messages.
// Returns raw event bytes (including 8-byte discriminator).
// Deduplicates events because emit_cpi! writes the same data twice in logs.
// Caller uses generated ParseAnyEvent() / ParseEvent_XxxEvent() to filter and decode.
func extractAnchorEvents(logs []string) [][]byte {
	var events [][]byte
	seen := make(map[string]struct{})
	for _, log := range logs {
		if !strings.HasPrefix(log, logPrefixProgramData) {
			continue
		}
		encoded := log[len(logPrefixProgramData):]
		if _, dup := seen[encoded]; dup {
			continue
		}
		seen[encoded] = struct{}{}
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			continue
		}
		events = append(events, data)
	}
	return events
}

// buildMintMap builds a mapping from token account address → mint address
// using Pre/PostTokenBalances. Shared by all handlers for token identification.
func buildMintMap(meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice) map[string]string {
	mintMap := make(map[string]string)
	for _, bal := range meta.PreTokenBalances {
		if int(bal.AccountIndex) >= len(accountKeys) {
			continue
		}
		account := accountKeys[bal.AccountIndex].String()
		mintMap[account] = bal.Mint.String()
	}
	for _, bal := range meta.PostTokenBalances {
		if int(bal.AccountIndex) >= len(accountKeys) {
			continue
		}
		account := accountKeys[bal.AccountIndex].String()
		if _, exists := mintMap[account]; !exists {
			mintMap[account] = bal.Mint.String()
		}
	}
	return mintMap
}

