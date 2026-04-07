package block

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	amm "splash.xyz/dex/pkg/contracts/pumpfun/amm/idl/generated"
	bonding_curve "splash.xyz/dex/pkg/contracts/pumpfun/bondingcurve/idl/generated"
)

// PumpFun Bonding Curve token decimals (always 6 for tokens created via pump.fun)
const pumpFunTokenDecimals = 6

func (s *BlockService) handlePumpFunBondingCurve(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ix solana.CompiledInstruction, ixIndex int) {
	_ = ctx
	_ = meta
	_ = ixIndex

	if err := decodePumpFunBondingCurveInstruction(txHash, ix, accountKeys); err != nil {
		fmt.Printf("[PumpFun BC IX] decode failed tx: %s, err: %v\n", txHash, err)
	}
}

// handlePumpFunBondingCurve handles swap instructions from the PumpFun Bonding Curve program.
// Uses Anchor Event parsing (TradeEvent) from LogMessages.
func (s *BlockService) handlePumpFunBondingCurveLogMessages(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	// TODO: LogMessages are transaction-scoped, but this handler is invoked per instruction.
	// A tx with multiple PumpFun instructions can re-consume the same events and print duplicates.
	// Revisit with a tx-scoped event cursor/key-counter approach before treating this as production-safe.
	// TODO: ixIndex is currently unused because events are not yet matched back to a specific instruction.
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
func (s *BlockService) handlePumpSwapAMMLogMessages(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ixIndex int) {
	// TODO: LogMessages are transaction-scoped, but this handler is invoked per instruction.
	// A tx with multiple PumpSwap instructions can re-consume the same events and print duplicates.
	// Revisit with a tx-scoped event cache plus key counters (for example txHash+ixName+pool) before production use.
	// TODO: ixIndex is currently unused because events are not yet matched back to a specific instruction.
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

func (s *BlockService) handlePumpSwapAMM(ctx context.Context, txHash string, meta *rpc.TransactionMeta, accountKeys solana.PublicKeySlice, ix solana.CompiledInstruction, ixIndex int) {
	_ = ctx
	_ = meta
	_ = ixIndex

	if err := decodePumpSwapAMMInstruction(txHash, ix, accountKeys); err != nil {
		fmt.Printf("[PumpSwap IX] decode failed tx: %s, err: %v\n", txHash, err)
	}
}

func decodePumpSwapAMMInstruction(txHash string, ix solana.CompiledInstruction, accountKeys solana.PublicKeySlice) error {
	ixName := decodePumpAmmInstructionName(ix)
	if ixName == "" {
		return fmt.Errorf("unknown pump amm instruction")
	}
	if err := validatePumpAmmAccounts(ixName, len(ix.Accounts)); err != nil {
		return fmt.Errorf("invalid accounts length: %w", err)
	}

	switch ixName {
	case "Buy", "BuyExactQuoteIn":
		pool := accountKeys[ix.Accounts[0]]
		user := accountKeys[ix.Accounts[1]]
		globalConfig := accountKeys[ix.Accounts[2]]
		baseMint := accountKeys[ix.Accounts[3]]
		quoteMint := accountKeys[ix.Accounts[4]]
		userBaseTokenAccount := accountKeys[ix.Accounts[5]]
		userQuoteTokenAccount := accountKeys[ix.Accounts[6]]
		poolBaseTokenAccount := accountKeys[ix.Accounts[7]]
		poolQuoteTokenAccount := accountKeys[ix.Accounts[8]]
		protocolFeeRecipient := accountKeys[ix.Accounts[9]]
		protocolFeeRecipientTokenAccount := accountKeys[ix.Accounts[10]]
		baseTokenProgram := accountKeys[ix.Accounts[11]]
		quoteTokenProgram := accountKeys[ix.Accounts[12]]
		systemProgram := accountKeys[ix.Accounts[13]]
		associatedTokenProgram := accountKeys[ix.Accounts[14]]
		eventAuthority := accountKeys[ix.Accounts[15]]
		program := accountKeys[ix.Accounts[16]]
		coinCreatorVaultAta := accountKeys[ix.Accounts[17]]
		coinCreatorVaultAuthority := accountKeys[ix.Accounts[18]]
		globalVolumeAccumulator := accountKeys[ix.Accounts[19]]
		userVolumeAccumulator := accountKeys[ix.Accounts[20]]
		feeConfig := accountKeys[ix.Accounts[21]]
		feeProgram := accountKeys[ix.Accounts[22]]

		fmt.Printf("[PumpSwap IX] %s tx: %s, pool: %s, user: %s, globalConfig: %s, baseMint: %s, quoteMint: %s, userBaseTokenAccount: %s, userQuoteTokenAccount: %s, poolBaseTokenAccount: %s, poolQuoteTokenAccount: %s, protocolFeeRecipient: %s, protocolFeeRecipientTokenAccount: %s, baseTokenProgram: %s, quoteTokenProgram: %s, systemProgram: %s, associatedTokenProgram: %s, eventAuthority: %s, program: %s, coinCreatorVaultAta: %s, coinCreatorVaultAuthority: %s, globalVolumeAccumulator: %s, userVolumeAccumulator: %s, feeConfig: %s, feeProgram: %s\n",
			ixName, txHash, pool, user, globalConfig, baseMint, quoteMint, userBaseTokenAccount, userQuoteTokenAccount, poolBaseTokenAccount, poolQuoteTokenAccount, protocolFeeRecipient, protocolFeeRecipientTokenAccount, baseTokenProgram, quoteTokenProgram, systemProgram, associatedTokenProgram, eventAuthority, program, coinCreatorVaultAta, coinCreatorVaultAuthority, globalVolumeAccumulator, userVolumeAccumulator, feeConfig, feeProgram)
		return nil
	case "Sell":
		pool := accountKeys[ix.Accounts[0]]
		user := accountKeys[ix.Accounts[1]]
		globalConfig := accountKeys[ix.Accounts[2]]
		baseMint := accountKeys[ix.Accounts[3]]
		quoteMint := accountKeys[ix.Accounts[4]]
		userBaseTokenAccount := accountKeys[ix.Accounts[5]]
		userQuoteTokenAccount := accountKeys[ix.Accounts[6]]
		poolBaseTokenAccount := accountKeys[ix.Accounts[7]]
		poolQuoteTokenAccount := accountKeys[ix.Accounts[8]]
		protocolFeeRecipient := accountKeys[ix.Accounts[9]]
		protocolFeeRecipientTokenAccount := accountKeys[ix.Accounts[10]]
		baseTokenProgram := accountKeys[ix.Accounts[11]]
		quoteTokenProgram := accountKeys[ix.Accounts[12]]
		systemProgram := accountKeys[ix.Accounts[13]]
		associatedTokenProgram := accountKeys[ix.Accounts[14]]
		eventAuthority := accountKeys[ix.Accounts[15]]
		program := accountKeys[ix.Accounts[16]]
		coinCreatorVaultAta := accountKeys[ix.Accounts[17]]
		coinCreatorVaultAuthority := accountKeys[ix.Accounts[18]]
		feeConfig := accountKeys[ix.Accounts[19]]
		feeProgram := accountKeys[ix.Accounts[20]]

		fmt.Printf("[PumpSwap IX] %s tx: %s, pool: %s, user: %s, globalConfig: %s, baseMint: %s, quoteMint: %s, userBaseTokenAccount: %s, userQuoteTokenAccount: %s, poolBaseTokenAccount: %s, poolQuoteTokenAccount: %s, protocolFeeRecipient: %s, protocolFeeRecipientTokenAccount: %s, baseTokenProgram: %s, quoteTokenProgram: %s, systemProgram: %s, associatedTokenProgram: %s, eventAuthority: %s, program: %s, coinCreatorVaultAta: %s, coinCreatorVaultAuthority: %s, feeConfig: %s, feeProgram: %s\n",
			ixName, txHash, pool, user, globalConfig, baseMint, quoteMint, userBaseTokenAccount, userQuoteTokenAccount, poolBaseTokenAccount, poolQuoteTokenAccount, protocolFeeRecipient, protocolFeeRecipientTokenAccount, baseTokenProgram, quoteTokenProgram, systemProgram, associatedTokenProgram, eventAuthority, program, coinCreatorVaultAta, coinCreatorVaultAuthority, feeConfig, feeProgram)
		return nil
	case "CreatePool":
		pool := accountKeys[ix.Accounts[0]]
		globalConfig := accountKeys[ix.Accounts[1]]
		creator := accountKeys[ix.Accounts[2]]
		baseMint := accountKeys[ix.Accounts[3]]
		quoteMint := accountKeys[ix.Accounts[4]]
		lpMint := accountKeys[ix.Accounts[5]]
		userBaseTokenAccount := accountKeys[ix.Accounts[6]]
		userQuoteTokenAccount := accountKeys[ix.Accounts[7]]
		userPoolTokenAccount := accountKeys[ix.Accounts[8]]
		poolBaseTokenAccount := accountKeys[ix.Accounts[9]]
		poolQuoteTokenAccount := accountKeys[ix.Accounts[10]]
		systemProgram := accountKeys[ix.Accounts[11]]
		token2022Program := accountKeys[ix.Accounts[12]]
		baseTokenProgram := accountKeys[ix.Accounts[13]]
		quoteTokenProgram := accountKeys[ix.Accounts[14]]
		associatedTokenProgram := accountKeys[ix.Accounts[15]]
		eventAuthority := accountKeys[ix.Accounts[16]]
		program := accountKeys[ix.Accounts[17]]

		fmt.Printf("[PumpSwap IX] %s tx: %s, pool: %s, globalConfig: %s, creator: %s, baseMint: %s, quoteMint: %s, lpMint: %s, userBaseTokenAccount: %s, userQuoteTokenAccount: %s, userPoolTokenAccount: %s, poolBaseTokenAccount: %s, poolQuoteTokenAccount: %s, systemProgram: %s, token2022Program: %s, baseTokenProgram: %s, quoteTokenProgram: %s, associatedTokenProgram: %s, eventAuthority: %s, program: %s\n",
			ixName, txHash, pool, globalConfig, creator, baseMint, quoteMint, lpMint, userBaseTokenAccount, userQuoteTokenAccount, userPoolTokenAccount, poolBaseTokenAccount, poolQuoteTokenAccount, systemProgram, token2022Program, baseTokenProgram, quoteTokenProgram, associatedTokenProgram, eventAuthority, program)
		return nil
	}

	return nil
}

func decodePumpFunBondingCurveInstruction(txHash string, ix solana.CompiledInstruction, accountKeys solana.PublicKeySlice) error {
	ixName := decodePumpFunInstruction(ix)
	if ixName == "" {
		return fmt.Errorf("unknown pumpfun bonding curve instruction")
	}
	if err := validatePumpFunBondingCurveAccounts(ixName, len(ix.Accounts)); err != nil {
		return fmt.Errorf("invalid accounts length: %w", err)
	}

	switch ixName {
	case "Buy", "BuyExactSolIn":
		global := accountKeys[ix.Accounts[0]]
		feeRecipient := accountKeys[ix.Accounts[1]]
		mint := accountKeys[ix.Accounts[2]]
		bondingCurve := accountKeys[ix.Accounts[3]]
		associatedBondingCurve := accountKeys[ix.Accounts[4]]
		associatedUser := accountKeys[ix.Accounts[5]]
		user := accountKeys[ix.Accounts[6]]
		systemProgram := accountKeys[ix.Accounts[7]]
		tokenProgram := accountKeys[ix.Accounts[8]]
		creatorVault := accountKeys[ix.Accounts[9]]
		eventAuthority := accountKeys[ix.Accounts[10]]
		program := accountKeys[ix.Accounts[11]]
		globalVolumeAccumulator := accountKeys[ix.Accounts[12]]
		userVolumeAccumulator := accountKeys[ix.Accounts[13]]
		feeConfig := accountKeys[ix.Accounts[14]]
		feeProgram := accountKeys[ix.Accounts[15]]

		fmt.Printf("[PumpFun BC IX] %s tx: %s, global: %s, feeRecipient: %s, mint: %s, bondingCurve: %s, associatedBondingCurve: %s, associatedUser: %s, user: %s, systemProgram: %s, tokenProgram: %s, creatorVault: %s, eventAuthority: %s, program: %s, globalVolumeAccumulator: %s, userVolumeAccumulator: %s, feeConfig: %s, feeProgram: %s\n",
			ixName, txHash, global, feeRecipient, mint, bondingCurve, associatedBondingCurve, associatedUser, user, systemProgram, tokenProgram, creatorVault, eventAuthority, program, globalVolumeAccumulator, userVolumeAccumulator, feeConfig, feeProgram)
		return nil
	case "Sell":
		global := accountKeys[ix.Accounts[0]]
		feeRecipient := accountKeys[ix.Accounts[1]]
		mint := accountKeys[ix.Accounts[2]]
		bondingCurve := accountKeys[ix.Accounts[3]]
		associatedBondingCurve := accountKeys[ix.Accounts[4]]
		associatedUser := accountKeys[ix.Accounts[5]]
		user := accountKeys[ix.Accounts[6]]
		systemProgram := accountKeys[ix.Accounts[7]]
		creatorVault := accountKeys[ix.Accounts[8]]
		tokenProgram := accountKeys[ix.Accounts[9]]
		eventAuthority := accountKeys[ix.Accounts[10]]
		program := accountKeys[ix.Accounts[11]]
		feeConfig := accountKeys[ix.Accounts[12]]
		feeProgram := accountKeys[ix.Accounts[13]]

		fmt.Printf("[PumpFun BC IX] %s tx: %s, global: %s, feeRecipient: %s, mint: %s, bondingCurve: %s, associatedBondingCurve: %s, associatedUser: %s, user: %s, systemProgram: %s, creatorVault: %s, tokenProgram: %s, eventAuthority: %s, program: %s, feeConfig: %s, feeProgram: %s\n",
			ixName, txHash, global, feeRecipient, mint, bondingCurve, associatedBondingCurve, associatedUser, user, systemProgram, creatorVault, tokenProgram, eventAuthority, program, feeConfig, feeProgram)
		return nil
	case "Create":
		mint := accountKeys[ix.Accounts[0]]
		mintAuthority := accountKeys[ix.Accounts[1]]
		bondingCurve := accountKeys[ix.Accounts[2]]
		associatedBondingCurve := accountKeys[ix.Accounts[3]]
		global := accountKeys[ix.Accounts[4]]
		mplTokenMetadata := accountKeys[ix.Accounts[5]]
		metadata := accountKeys[ix.Accounts[6]]
		user := accountKeys[ix.Accounts[7]]
		systemProgram := accountKeys[ix.Accounts[8]]
		tokenProgram := accountKeys[ix.Accounts[9]]
		associatedTokenProgram := accountKeys[ix.Accounts[10]]
		rent := accountKeys[ix.Accounts[11]]
		eventAuthority := accountKeys[ix.Accounts[12]]
		program := accountKeys[ix.Accounts[13]]

		fmt.Printf("[PumpFun BC IX] %s tx: %s, mint: %s, mintAuthority: %s, bondingCurve: %s, associatedBondingCurve: %s, global: %s, mplTokenMetadata: %s, metadata: %s, user: %s, systemProgram: %s, tokenProgram: %s, associatedTokenProgram: %s, rent: %s, eventAuthority: %s, program: %s\n",
			ixName, txHash, mint, mintAuthority, bondingCurve, associatedBondingCurve, global, mplTokenMetadata, metadata, user, systemProgram, tokenProgram, associatedTokenProgram, rent, eventAuthority, program)
		return nil
	case "CreateV2":
		mint := accountKeys[ix.Accounts[0]]
		mintAuthority := accountKeys[ix.Accounts[1]]
		bondingCurve := accountKeys[ix.Accounts[2]]
		associatedBondingCurve := accountKeys[ix.Accounts[3]]
		global := accountKeys[ix.Accounts[4]]
		user := accountKeys[ix.Accounts[5]]
		systemProgram := accountKeys[ix.Accounts[6]]
		tokenProgram := accountKeys[ix.Accounts[7]]
		associatedTokenProgram := accountKeys[ix.Accounts[8]]
		mayhemProgramID := accountKeys[ix.Accounts[9]]
		globalParams := accountKeys[ix.Accounts[10]]
		solVault := accountKeys[ix.Accounts[11]]
		mayhemState := accountKeys[ix.Accounts[12]]
		mayhemTokenVault := accountKeys[ix.Accounts[13]]
		eventAuthority := accountKeys[ix.Accounts[14]]
		program := accountKeys[ix.Accounts[15]]

		fmt.Printf("[PumpFun BC IX] %s tx: %s, mint: %s, mintAuthority: %s, bondingCurve: %s, associatedBondingCurve: %s, global: %s, user: %s, systemProgram: %s, tokenProgram: %s, associatedTokenProgram: %s, mayhemProgramID: %s, globalParams: %s, solVault: %s, mayhemState: %s, mayhemTokenVault: %s, eventAuthority: %s, program: %s\n",
			ixName, txHash, mint, mintAuthority, bondingCurve, associatedBondingCurve, global, user, systemProgram, tokenProgram, associatedTokenProgram, mayhemProgramID, globalParams, solVault, mayhemState, mayhemTokenVault, eventAuthority, program)
		return nil
	}

	return nil
}

func validatePumpAmmAccounts(ixName string, n int) error {

	switch ixName {
	case "Buy", "BuyExactQuoteIn":
		if n != 23 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	case "Sell":
		if n != 21 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	case "CreatePool":
		if n != 18 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	}
	return nil
}

func validatePumpFunBondingCurveAccounts(ixName string, n int) error {
	switch ixName {
	case "Buy", "BuyExactSolIn":
		if n != 16 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	case "Sell":
		if n != 14 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	case "Create":
		if n != 14 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	case "CreateV2":
		if n != 16 {
			return fmt.Errorf("%s invalid accounts length: %d", ixName, n)
		}
	}
	return nil
}

func decodePumpAmmInstructionName(ix solana.CompiledInstruction) string {
	if len(ix.Data) < 8 {
		return ""
	}

	switch {
	case bytes.Equal(ix.Data[:8], amm.Instruction_Buy[:]):
		return "Buy"
	case bytes.Equal(ix.Data[:8], amm.Instruction_Sell[:]):
		return "Sell"
	case bytes.Equal(ix.Data[:8], amm.Instruction_BuyExactQuoteIn[:]):
		return "BuyExactQuoteIn"
	case bytes.Equal(ix.Data[:8], amm.Instruction_CreatePool[:]):
		return "CreatePool"
	case bytes.Equal(ix.Data[:8], amm.Instruction_CreateConfig[:]):
		return "CreateConfig"
	case bytes.Equal(ix.Data[:8], amm.Instruction_Deposit[:]):
		return "Deposit"
	case bytes.Equal(ix.Data[:8], amm.Instruction_Disable[:]):
		return "Disable"
	case bytes.Equal(ix.Data[:8], amm.Instruction_ExtendAccount[:]):
		return "ExtendAccount"
	case bytes.Equal(ix.Data[:8], amm.Instruction_InitUserVolumeAccumulator[:]):
		return "InitUserVolumeAccumulator"
	case bytes.Equal(ix.Data[:8], amm.Instruction_MigratePoolCoinCreator[:]):
		return "MigratePoolCoinCreator"
	case bytes.Equal(ix.Data[:8], amm.Instruction_SyncUserVolumeAccumulator[:]):
		return "SyncUserVolumeAccumulator"
	case bytes.Equal(ix.Data[:8], amm.Instruction_ToggleCashbackEnabled[:]):
		return "ToggleCashbackEnabled"
	case bytes.Equal(ix.Data[:8], amm.Instruction_ToggleMayhemMode[:]):
		return "ToggleMayhemMode"
	case bytes.Equal(ix.Data[:8], amm.Instruction_TransferCreatorFeesToPump[:]):
		return "TransferCreatorFeesToPump"
	case bytes.Equal(ix.Data[:8], amm.Instruction_UpdateAdmin[:]):
		return "UpdateAdmin"
	case bytes.Equal(ix.Data[:8], amm.Instruction_UpdateFeeConfig[:]):
		return "UpdateFeeConfig"
	case bytes.Equal(ix.Data[:8], amm.Instruction_Withdraw[:]):
		return "Withdraw"
	default:
		return ""
	}
}

func decodePumpFunInstruction(ix solana.CompiledInstruction) string {
	if len(ix.Data) < 8 {
		return ""
	}

	switch {
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_Create[:]):
		return "Create"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_Buy[:]):
		return "Buy"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_BuyExactSolIn[:]):
		return "BuyExactSolIn"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_Sell[:]):
		return "Sell"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_CreateV2[:]):
		return "CreateV2"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_Migrate[:]):
		return "Migrate"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_MigrateBondingCurveCreator[:]):
		return "MigrateBondingCurveCreator"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SetCreator[:]):
		return "SetCreator"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SetMayhemVirtualParams[:]):
		return "SetMayhemVirtualParams"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SetMetaplexCreator[:]):
		return "SetMetaplexCreator"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SetParams[:]):
		return "SetParams"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SetReservedFeeRecipients[:]):
		return "SetReservedFeeRecipients"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_SyncUserVolumeAccumulator[:]):
		return "SyncUserVolumeAccumulator"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_ToggleCashbackEnabled[:]):
		return "ToggleCashbackEnabled"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_ToggleCreateV2[:]):
		return "ToggleCreateV2"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_ToggleMayhemMode[:]):
		return "ToggleMayhemMode"
	case bytes.Equal(ix.Data[:8], bonding_curve.Instruction_UpdateGlobalAuthority[:]):
		return "UpdateGlobalAuthority"
	default:
		return ""
	}
}
