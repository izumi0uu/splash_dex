package block

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go/rpc"

	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"splash.xyz/dex/consumer/internal/config"
	"splash.xyz/dex/consumer/internal/model"
	"splash.xyz/dex/consumer/internal/svc"
	"splash.xyz/dex/pkg/constants"
)

type BlockService struct {
	c  *rpc.Client
	sc *svc.ServiceContext
	logx.Logger
	workerPool *ants.Pool
	slotChan   chan uint64
	failedChan chan uint64
	solPrice   float64
	slot       uint64
	Conn       *websocket.Conn
	ctx        context.Context
	cancel     func(err error)
	name       string
}

func (s *BlockService) Stop() {
	s.cancel(constants.ErrServiceStop)
	if s.Conn != nil {
		err := s.Conn.WriteMessage(websocket.TextMessage, []byte("{\"id\":1,\"jsonrpc\":\"2.0\",\"method\": \"blockUnsubscribe\", \"params\": [0]}\n"))
		if err != nil {
			s.Error("programUnsubscribe", err)
		}
		_ = s.Conn.Close()
	}
}

func (s *BlockService) Start() {
	s.GetBlockFromHttp()
}

func NewBlockService(sc *svc.ServiceContext, name string, slotChan chan uint64, failedChan chan uint64, index int) *BlockService {
	ctx, cancel := context.WithCancelCause(context.Background())
	pool, _ := ants.NewPool(5)
	solService := &BlockService{
		c:          svc.NewSolRPCClient(config.FindChainRpcByChainId(constants.SolChainIdInt)),
		sc:         sc,
		Logger:     logx.WithContext(context.Background()).WithFields(logx.Field("service", fmt.Sprintf("%s-%v", name, index))),
		slotChan:   slotChan,
		failedChan: failedChan,
		workerPool: pool,
		ctx:        ctx,
		cancel:     cancel,
		name:       name,
	}
	return solService
}

func (s *BlockService) GetBlockFromHttp() {
	ctx := s.ctx
	for {
		select {
		case <-s.ctx.Done():
			return
		case slot, ok := <-s.slotChan:
			if !ok {
				return
			}
			// fmt.Println("current slot is:", slot)
			threading.RunSafe(func() {
				s.ProcessBlock(ctx, int64(slot))
			})
		}
	}
}

func (s *BlockService) ProcessBlock(ctx context.Context, slot int64) {
	if slot == 0 {
		return
	}

	blockRecord := &model.SolBlocks{
		ChainId: s.sc.Config.Sol.ChainId,
		Slot:    uint64(slot),
		Status:  constants.BlockPending,
		ErrMsg:  NullableString(""),
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			blockRecord.Status = constants.BlockFailed
			blockRecord.ErrMsg = TrimmedNullableString(fmt.Sprintf("panic: %v", recovered))
			s.Errorf("process block panic, slot=%d, err=%v", slot, recovered)
		}
		if blockRecord.Status == constants.BlockPending {
			blockRecord.Status = constants.BlockFailed
		}
		s.persistBlock(blockRecord)
	}()

	blockInfo, err := GetSolBlockInfoDelay(s.sc.GetSolClient(), ctx, uint64(slot))
	if err != nil || blockInfo == nil {
		if err == nil {
			err = fmt.Errorf("empty block info")
		}
		blockRecord.Status = classifyBlockStatus(err)
		blockRecord.ErrMsg = TrimmedNullableString(err.Error())
		s.Errorf("get block info error, slot=%d, err=%v", slot, err)
		s.enqueueFailedSlot(uint64(slot), blockRecord.Status)
		return
	}

	if blockInfo.BlockHeight != nil {
		blockRecord.BlockHeight = NullableInt64(int64(*blockInfo.BlockHeight))
	}
	blockRecord.ParentSlot = NullableInt64(int64(blockInfo.ParentSlot))
	blockRecord.BlockHash = NullableString(blockInfo.Blockhash.String())
	blockRecord.PreviousBlockHash = NullableString(blockInfo.PreviousBlockhash.String())
	blockRecord.TxCount = int64(len(blockInfo.Transactions))
	if blockInfo.BlockTime != nil {
		blockRecord.BlockTime = NullableTime(blockInfo.BlockTime.Time())
	}

	for txIdx := range blockInfo.Transactions {
		txWithMeta := &blockInfo.Transactions[txIdx]

		// guard 1: skip failed transaction
		if txWithMeta.Meta == nil || txWithMeta.Meta.Err != nil {
			continue
		}

		// decode transaction from base64
		parsedTx, err := txWithMeta.GetTransaction()
		if err != nil || parsedTx == nil {
			continue
		}

		// guard 2: skip non-signature transaction
		if len(parsedTx.Signatures) == 0 {
			continue
		}

		// Resolve all account keys (static + loaded addresses from ALT)
		accountKeys := parsedTx.Message.AccountKeys
		if txWithMeta.Meta.LoadedAddresses.Writable != nil {
			accountKeys = append(accountKeys, txWithMeta.Meta.LoadedAddresses.Writable...)
		}
		if txWithMeta.Meta.LoadedAddresses.ReadOnly != nil {
			accountKeys = append(accountKeys, txWithMeta.Meta.LoadedAddresses.ReadOnly...)
		}

		// guard 3: skip vote transaction (~80% of all txs)
		isVote := true
		for _, instruction := range parsedTx.Message.Instructions {
			if int(instruction.ProgramIDIndex) >= len(accountKeys) {
				continue
			}
			program := accountKeys[instruction.ProgramIDIndex].String()
			if program != constants.ProgramStrVote {
				isVote = false
				break
			}
		}
		if isVote {
			continue
		}

		txHash := parsedTx.Signatures[0].String()

		for i, ix := range parsedTx.Message.Instructions {
			if int(ix.ProgramIDIndex) >= len(accountKeys) {
				continue
			}
			program := accountKeys[ix.ProgramIDIndex].String()

			switch program {
			// --- DEX ---
			case constants.ProgramStrRaydiumV4AMM:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "RaydiumV4")
			case constants.ProgramStrRaydiumV4CLMM:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "RaydiumCLMM")
			case constants.ProgramStrRaydiumCPMM:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "RaydiumCPMM")
			case constants.ProgramStrRaydiumV2:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "RaydiumV2")
			case constants.ProgramStrOrca:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "Orca")
			case constants.ProgramStrMeteoraDLMM:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "MeteoraDLMM")
			case constants.ProgramStrMeteoraPool:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "MeteoraPool")
			case constants.ProgramStrPhoenix:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "Phoenix")
			case constants.ProgramStrLifinity:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "Lifinity")
			// --- Launchpad ---
			case constants.ProgramStrPumpFun:
				s.handlePumpFunBondingCurve(ctx, txHash, txWithMeta.Meta, accountKeys, ix, i)
			case constants.ProgramStrPumpAmm:
				s.handlePumpSwapAMM(ctx, txHash, txWithMeta.Meta, accountKeys, ix, i)
			case constants.ProgramStrMoonshot:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "Moonshot")
			// --- Aggregator ---
			case constants.ProgramStrJupiter:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "Jupiter")
			default:
				continue
			}
		}
	}

	blockRecord.Status = constants.BlockProcessed
	blockRecord.ErrMsg = NullableString("")
}

func (s *BlockService) persistBlock(blockRecord *model.SolBlocks) {
	if blockRecord == nil || s.sc.SolBlockModel == nil {
		return
	}

	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.sc.SolBlockModel.Upsert(writeCtx, blockRecord)
	if err != nil {
		s.Errorf("upsert sol block error, slot=%d, err=%v", blockRecord.Slot, err)
	}
}

func classifyBlockStatus(err error) int64 {
	if err == nil {
		return constants.BlockProcessed
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "was skipped") ||
		strings.Contains(errMsg, "ledger jump") ||
		strings.Contains(errMsg, "slot was skipped") {
		return constants.BlockSkipped
	}

	return constants.BlockFailed
}

func (s *BlockService) enqueueFailedSlot(slot uint64, status int64) {
	if s.failedChan == nil || slot == 0 {
		return
	}

	// Skipped slots are deterministic and don't benefit from in-memory retry.
	if status == constants.BlockSkipped {
		return
	}

	select {
	case <-s.ctx.Done():
		return
	case s.failedChan <- slot:
		s.Infof("enqueue failed slot for retry, slot=%d", slot)
	default:
		s.Errorf("failed slot queue is full, drop slot=%d", slot)
	}
}
