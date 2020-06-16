package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.setram", &OpSetRAM{})
}

type OpSetRAM struct {
	MaxRAMSize uint64 `json:"max_ram_size"`
}

func (op *OpSetRAM) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	return append(out, system.NewSetRAM(op.MaxRAMSize)), nil
}
