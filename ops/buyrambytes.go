package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
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

func (op *OpBuyRamBytes) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewBuyRAMBytes(op.Payer, op.Receiver, op.Bytes))
	in <- EndTransaction(opPubkey) // end transaction
	return nil

}