package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.buy_ram_bytes", &OpBuyRamBytes{})
}

type OpBuyRamBytes struct {
	Payer    eos.AccountName
	Receiver eos.AccountName
	Bytes    uint32
}

func (op *OpBuyRamBytes) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	return append(out, system.NewBuyRAMBytes(op.Payer, op.Receiver, op.Bytes)), nil
}
