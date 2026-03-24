package slot

import (
	"errors"

	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/threading"
)

type SlotWsService struct {
	*SlotService
}

func NewSlotWsService(slotService *SlotService) *SlotWsService {
	return &SlotWsService{
		SlotService: slotService,
	}
}

func (s *SlotWsService) Start() {
	// register shutdown listener for graceful cleanup
	proc.AddShutdownListener(func() {
		s.Infof("SlotWsService:ShutdownListener")
		s.cancel(errors.New("close slot"))
	})

	s.SlotWs()
}

func (s *SlotWsService) SlotWs() {
	s.MustConnect()
	s.Infof("SlotWs:MustConnect success")

	// safe goroutine with panic recovery
	threading.GoSafe(func() {
		for {
			select {
			case <-s.ctx.Done():
				s.Info("slotWs stop succeed")
				return
			default:
			}
			s.ReadSlotMessage()
		}
	})
}

func (s *SlotWsService) Stop() {
	s.Info("stopping slot websocket service")
}
