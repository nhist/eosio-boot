package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	Register("token.transfer", &OpTransferToken{})
}

type OpTransferToken struct {
	From     eos.AccountName
	To       eos.AccountName
	Quantity eos.Asset
	Memo     string
}

func (op *OpTransferToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewTransfer(op.From, op.To, op.Quantity, op.Memo))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
