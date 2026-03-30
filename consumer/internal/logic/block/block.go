package block

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go/rpc"

	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"splash.xyz/dex/consumer/internal/config"
	"splash.xyz/dex/consumer/internal/svc"
	"splash.xyz/dex/pkg/constants"
)

type BlockService struct {
	c  *rpc.Client
	sc *svc.ServiceContext
	logx.Logger
	workerPool *ants.Pool
	slotChan   chan uint64
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

func NewBlockService(sc *svc.ServiceContext, name string, slotChan chan uint64, index int) *BlockService {
	ctx, cancel := context.WithCancelCause(context.Background())
	pool, _ := ants.NewPool(5)
	solService := &BlockService{
		c:          svc.NewSolRPCClient(config.FindChainRpcByChainId(constants.SolChainIdInt)),
		sc:         sc,
		Logger:     logx.WithContext(context.Background()).WithFields(logx.Field("service", fmt.Sprintf("%s-%v", name, index))),
		slotChan:   slotChan,
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

	blockInfo, err := GetSolBlockInfoDelay(s.sc.GetSolClient(), ctx, uint64(slot))
	if err != nil || blockInfo == nil {
		fmt.Println("get block info error", err)
		return
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
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "PumpFun")
			case constants.ProgramStrPumpAmm:
				s.handleDexSwap(ctx, txHash, txWithMeta.Meta, accountKeys, i, "PumpSwap")
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
}
