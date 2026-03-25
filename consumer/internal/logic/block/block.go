package block

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gorilla/websocket"
	"github.com/mr-tron/base58"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"splash.xyz/dex/consumer/internal/config"
	"splash.xyz/dex/consumer/internal/svc"
	"splash.xyz/dex/pkg/constants"
)

type BlockService struct {
	c  *client.Client
	sc *svc.ServiceContext
	logx.Logger
	workerPool *ants.Pool
	slotChan   chan uint64
	// holders    *datastructure.CopyOnWriteList[*tokenpkg.ProgramAccount]
	solPrice float64
	slot     uint64
	Conn     *websocket.Conn
	ctx      context.Context
	cancel   func(err error)
	name     string
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
		c: client.New(rpc.WithEndpoint(config.FindChainRpcByChainId(constants.SolChainIdInt)), rpc.WithHTTPClient(&http.Client{
			Timeout: 5 * time.Second,
		})),
		// holders:    datastructure.NewCopyOnWriteList[*tokenpkg.ProgramAccount]([]*tokenpkg.ProgramAccount{}),
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

	slice.ForEach(blockInfo.Transactions, func(index int, tx client.BlockTransaction) {
		if len(tx.Transaction.Signatures) > 0 {
			sig858 := base58.Encode(tx.Transaction.Signatures[0])
			fmt.Println("Transaction signature: ", sig858)
		}
	})

}
