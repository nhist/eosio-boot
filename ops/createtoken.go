package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	Register("token.create", &OpCreateToken{})
}

type OpCreateToken struct {
	// TODO: this should have be Issuer
	Account eos.AccountName `json:"account"`
	// TODO: this should be MaximumSupply
	Amount eos.Asset `json:"amount"`
}

func (op *OpCreateToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewCreate(op.Account, op.Amount))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func (op *OpCreateToken) RequireValidation() bool {
	return true
}
