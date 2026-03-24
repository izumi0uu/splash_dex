package slot

import (
	"splash.xyz/dex/consumer/internal/svc"
)

type SlotServiceGroup struct {
	*SlotService
	Ws *SlotWsService
}

func NewSlotServiceGroup(sc *svc.ServiceContext, slotChan chan uint64) *SlotServiceGroup {
	slotService := NewSlotService(sc, slotChan)

	return &SlotServiceGroup{
		SlotService: slotService,
		Ws:          NewSlotWsService(slotService),
	}
}

func (s *SlotServiceGroup) Start() {
	s.Ws.Start()
}

func (s *SlotServiceGroup) Stop() {
	s.Ws.Stop()
	s.SlotService.Stop()
}
