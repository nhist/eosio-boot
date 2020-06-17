package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.setpriv", &OpSetPriv{})
}


type OpSetPriv struct {
	Account eos.AccountName
}

func (op *OpSetPriv) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- system.NewSetPriv(op.Account)
	in <- EndTransaction(opPubkey) // end transaction
	return nil

}
