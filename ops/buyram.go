package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)


func init() {
	eosboot.Register("system.buy_ram", &OpBuyRam{})
}


type OpBuyRam struct {
	Payer       eos.AccountName
	Receiver    eos.AccountName
	EOSQuantity uint64 `json:"eos_quantity"`
}

func (op *OpBuyRam) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	return append(out, system.NewBuyRAM(op.Payer, op.Receiver, op.EOSQuantity)), nil
}

