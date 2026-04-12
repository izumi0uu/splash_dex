package slot

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"splash.xyz/dex/consumer/internal/model"
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

	s.scanAndEnqueue()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.scanAndEnqueue()
		}
	}
}

func (s *SlotNotCompleteService) Stop() {
	s.cancel(ErrServiceStop)
}

func (s *SlotNotCompleteService) scanAndEnqueue() {
	if s.sc.SolBlockModel == nil || s.errorChan == nil {
		return
	}

	queryCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	blocks, err := s.sc.SolBlockModel.FindRetryableSlots(queryCtx, s.sc.Config.Sol.ChainId, retryableSlotLimit)
	switch err {
	case nil:
	case model.ErrNotFound:
		return
	default:
		s.Errorf("scan retryable slots error: %v", err)
		return
	}

	now := time.Now()
	s.pruneEnqueued(now)

	for _, block := range blocks {
		if block == nil || block.Slot == 0 {
			continue
		}
		if last, ok := s.lastEnqueued[block.Slot]; ok && now.Sub(last) < retryableSlotCooldown {
			continue
		}

		select {
		case <-s.ctx.Done():
			return
		case s.errorChan <- block.Slot:
			s.lastEnqueued[block.Slot] = now
			s.Infof("enqueue retryable slot=%d status=%d", block.Slot, block.Status)
		default:
			s.Errorf("retryable slot queue is full, drop slot=%d", block.Slot)
			return
		}
	}
}

func (s *SlotNotCompleteService) pruneEnqueued(now time.Time) {
	if len(s.lastEnqueued) == 0 {
		return
	}

	for slot, at := range s.lastEnqueued {
		if now.Sub(at) > 10*retryableSlotCooldown {
			delete(s.lastEnqueued, slot)
		}
	}
}
