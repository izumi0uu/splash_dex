package slot

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"splash.xyz/dex/consumer/internal/svc"
)

const (
	retryableSlotLimit    = 50
	retryableScanInterval = 5 * time.Second
	retryableSlotCooldown = 30 * time.Second
)

type SlotNotCompleteService struct {
	sc *svc.ServiceContext
	logx.Logger

	ctx       context.Context
	cancel    context.CancelCauseFunc
	errorChan chan uint64

	lastEnqueued map[uint64]time.Time
}

func NewSlotNotCompleteService(sc *svc.ServiceContext, errorChan chan uint64) *SlotNotCompleteService {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &SlotNotCompleteService{
		sc:           sc,
		Logger:       logx.WithContext(ctx).WithFields(logx.Field("service", "slot-not-complete")),
		ctx:          ctx,
		cancel:       cancel,
		errorChan:    errorChan,
		lastEnqueued: make(map[uint64]time.Time),
	}
}

func (s *SlotNotCompleteService) Start() {
	ticker := time.NewTicker(retryableScanInterval)
	defer ticker.Stop()

	s.RecoverFailedBlock()
}

func (s *SlotNotCompleteService) Stop() {
	s.cancel(ErrServiceStop)
}

func (s *SlotNotCompleteService) RecoverFailedBlock() {
	slot := s.sc.Config.Sol.StartBlock

	if slot == 0 {
		block, err := s.sc.SolBlockModel.FindFirstFailBlock(s.ctx, s.sc.Config.Sol.ChainId)
		if err != nil {
			// no failed block or query error
			return
		}
		slot = block.Slot
	}

	checkTicker := time.NewTicker(5 * time.Second)
	sendTicker := time.NewTicker(5 * time.Second)
	defer checkTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-checkTicker.C:
			fromSlot := slot
			if fromSlot > 100 {
				fromSlot = fromSlot - 100
			}

			blocks, err := s.sc.SolBlockModel.FindProcessingSlots(
				s.ctx,
				s.sc.Config.Sol.ChainId,
				fromSlot,
				50,
			)
			if err != nil {
				continue
			}

			for _, block := range blocks {
				select {
				case <-s.ctx.Done():
					return
				case <-sendTicker.C:
					s.errorChan <- block.Slot
				}
			}
		}
	}
}
