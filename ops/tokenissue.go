package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
)


func init() {
	Register("token.issue", &OpIssueToken{})
}


type OpIssueToken struct {
	Account eos.AccountName
	Amount  eos.Asset
	Memo    string
}

func (op *OpIssueToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewIssue(op.Account, op.Amount, op.Memo))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
