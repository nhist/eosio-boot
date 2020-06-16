package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
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

func (op *OpIssueToken) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	act := token.NewIssue(op.Account, op.Amount, op.Memo)
	return append(out, act), nil
}
