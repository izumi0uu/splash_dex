package types

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
