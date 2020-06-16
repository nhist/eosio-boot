package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	eosboot.Register("token.create", &OpCreateToken{})
}


type OpCreateToken struct {
	Account eos.AccountName `json:"account"`
	Amount  eos.Asset       `json:"amount"`
}

func (op *OpCreateToken) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	act := token.NewCreate(op.Account, op.Amount)
	return append(out, act), nil
}
