package slot

import (
	"splash.xyz/dex/consumer/internal/svc"
)

type SlotServiceGroup struct {
	*SlotService
	Ws           *SlotWsService
	NotCompleted *SlotNotCompleteService // not-completed slot recovery
}

func NewSlotServiceGroup(sc *svc.ServiceContext, slotChan chan uint64, failedSlotChan chan uint64) *SlotServiceGroup {
	slotService := NewSlotService(sc, slotChan)

	return &SlotServiceGroup{
		SlotService:  slotService,
		Ws:           NewSlotWsService(slotService),
		NotCompleted: NewSlotNotCompleteService(sc, failedSlotChan),
	}
}

func (s *SlotServiceGroup) Start() {
	go s.NotCompleted.Start()
	s.Ws.Start()
}

func (s *SlotServiceGroup) Stop() {
	s.Ws.Stop()
	s.SlotService.Stop()
	s.NotCompleted.Stop()
}
