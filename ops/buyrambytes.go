package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	eosboot.Register("system.buy_ram_bytes", &OpBuyRamBytes{})
}

type OpBuyRamBytes struct {
	Payer    eos.AccountName
	Receiver eos.AccountName
	Bytes    uint32
}

func (op *OpBuyRamBytes) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	return append(out, system.NewBuyRAMBytes(op.Payer, op.Receiver, op.Bytes)), nil
}
