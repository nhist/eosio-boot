package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	eosboot.Register("system.setpriv", &OpSetPriv{})
}


type OpSetPriv struct {
	Account eos.AccountName
}

func (op *OpSetPriv) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	return append(out, system.NewSetPriv(op.Account)), nil
}
