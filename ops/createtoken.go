package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	Register("token.create", &OpCreateToken{})
}


type OpCreateToken struct {
	Account eos.AccountName `json:"account"`
	Amount  eos.Asset       `json:"amount"`
}

func (op *OpCreateToken) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	act := token.NewCreate(op.Account, op.Amount)
	return append(out, act), nil
}
