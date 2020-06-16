package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.setpriv", &OpSetPriv{})
}


type OpSetPriv struct {
	Account eos.AccountName
}

func (op *OpSetPriv) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	return append(out, system.NewSetPriv(op.Account)), nil
}
