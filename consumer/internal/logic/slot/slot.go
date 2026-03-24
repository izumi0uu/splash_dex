package slot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"splash.xyz/dex/consumer/internal/svc"
)

var ErrServiceStop = errors.New("service is stopped")

type SlotResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Slot   uint64 `json:"slot"`
			Parent uint64 `json:"parent"`
			Root   uint64 `json:"root"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

type SlotService struct {
	Conn *websocket.Conn
	sc   *svc.ServiceContext
	logx.Logger

	ctx     context.Context
	cancel  context.CancelCauseFunc
	maxSlot uint64

	realtimeChan chan uint64
}

func NewSlotService(sc *svc.ServiceContext, slotChan chan uint64) *SlotService {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &SlotService{
		Logger:       logx.WithContext(context.Background()).WithFields(logx.Field("service", "slot")),
		sc:           sc,
		ctx:          ctx,
		cancel:       cancel,
		realtimeChan: slotChan,
	}
}

// MustConnect connects to WebSocket and subscribes to slot updates with retry
func (s *SlotService) MustConnect() {
	dialer := websocket.DefaultDialer

	for {
		s.Infof("MustConnect: slot ws url: %v", s.sc.Config.Sol.WSUrl)
		dialer.HandshakeTimeout = time.Second * 5

		c, _, err := dialer.Dial(s.sc.Config.Sol.WSUrl, nil)
		if err != nil {
			s.Errorf("MustConnect: slot ws Dial err: %v", err)
		} else {
			s.Conn = c
			// try subscribe up to 10 times
			for i := 0; i < 10; i++ {
				err = c.WriteMessage(websocket.TextMessage, []byte(`{"id":1,"jsonrpc":"2.0","method":"slotSubscribe"}`))
				if err != nil {
					s.Error("slot ws slotSubscribe err: %v", err)
				} else {
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// ReadSlotMessage reads a single slot message with panic recovery and auto-reconnect
func (s *SlotService) ReadSlotMessage() {
	defer func() {
		cause := recover()
		if cause != nil {
			s.Error("ReadSlotMessage panic:", cause)
			s.MustConnect()
		}
	}()

	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		s.Error("SlotWs ReadMessage", err)
		// check if connection is closed or broken, auto-reconnect
		if strings.Contains(err.Error(), "close") {
			s.MustConnect()
		}
		if strings.Contains(err.Error(), "broken pipe") {
			s.MustConnect()
		}
		return
	}

	var resp SlotResp
	err = json.Unmarshal(message, &resp)
	if err != nil {
		s.Error("SlotWs json.Unmarshal", err)
		return
	}

	if resp.Params.Result.Slot == 0 {
		return
	}

	s.maxSlot = resp.Params.Result.Slot
	fmt.Println("latest slot: ", s.maxSlot)
	s.realtimeChan <- s.maxSlot
}

func (s *SlotService) Start() {
	// TODO: additional start logic
}

func (s *SlotService) Stop() {
	s.Info("stop slot service")
	s.cancel(ErrServiceStop)
	if s.Conn != nil {
		_ = s.Conn.Close()
	}
}
