package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	eosboot.Register("token.transfer", &OpTransferToken{})
}

type OpTransferToken struct {
	From     eos.AccountName
	To       eos.AccountName
	Quantity eos.Asset
	Memo     string
}

func (op *OpTransferToken) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	act := token.NewTransfer(op.From, op.To, op.Quantity, op.Memo)
	return append(out, act), nil
}
