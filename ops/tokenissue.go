package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)


func init() {
	eosboot.Register("token.issue", &OpIssueToken{})
}


type OpIssueToken struct {
	Account eos.AccountName
	Amount  eos.Asset
	Memo    string
}

func (op *OpIssueToken) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	act := token.NewIssue(op.Account, op.Amount, op.Memo)
	return append(out, act), nil
}
