package chain

import (
	"encoding/json"
	"splash.xyz/dex/consumer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

func SlotListener(client *Client, slotChan chan uint64) {
	if err := client.Connect(); err != nil {
		logx.Errorf("SlotListener: failed to connect: %v", err)
		return
	}

	subscribeMessage := `{"id":1,"jsonrpc":"2.0","method":"slotSubscribe"}`

	err := client.SendMessage([]byte(subscribeMessage))
	if err != nil {
		logx.Errorf("SlotListener: SendMessage error: %v", err)
		return
	}

	// subscription confirm
	confirmMessage, err := client.ReadMessage()
	if err != nil {
		logx.Errorf("SlotListener: ReadMessage error: %v", err)
		return
	}
	logx.Infof("SlotListener: confirmMessage: %s", string(confirmMessage))

	logx.Info("SlotListener: subscribed to slot updates")

	// read slot messages and send to channel
	for {
		message, err := client.ReadMessage()
		if err != nil {
			logx.Errorf("SlotListener: ReadMessage error: %v", err)
			return
		}

		var resp types.SlotResp
		err = json.Unmarshal(message, &resp)
		if err != nil {
			logx.Errorf("SlotListener: Unmarshal error: %v", err)
			continue // Continue reading instead of returning
		}

		// Only send valid slot numbers
		if resp.Params.Result.Slot > 0 {
			slotChan <- resp.Params.Result.Slot
			logx.Infof("SlotListener: received slot %d", resp.Params.Result.Slot)
		}
	}
}
