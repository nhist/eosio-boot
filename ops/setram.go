package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	eosboot.Register("system.setram", &OpSetRAM{})
}

type OpSetRAM struct {
	MaxRAMSize uint64 `json:"max_ram_size"`
}

func (op *OpSetRAM) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	return append(out, system.NewSetRAM(op.MaxRAMSize)), nil
}
