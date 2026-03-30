package block

import "github.com/gagliardetto/solana-go"

// TokenTransfer represents a parsed SPL Token transfer from inner instructions.
type TokenTransfer struct {
	From   solana.PublicKey
	To     solana.PublicKey
	Auth   solana.PublicKey
	Amount uint64
}

type TokenAccount struct {
	Owner               string
	TokenAccountAddress string
	TokenAddress        string
	TokenDecimal        uint8
	PreValue            int64
	PostValue           int64
	Closed              bool
	Init                bool
	PostValueUIString   string
	PreValueUIString    string
}

type Swap struct {
	BaseTokenInfo      *TokenAccount
	TokenInfo          *TokenAccount
	BaseTokenAmount    float64
	TokenAmount        float64
	BaseTokenAmountInt int64
	TokenAmountInt     int64
	Type               string
	To                 string
}
