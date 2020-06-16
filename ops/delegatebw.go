package boot


import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	eosboot.Register("system.delegate_bw", &OpDelegateBW{})
}


type OpDelegateBW struct {
	From     eos.AccountName
	To       eos.AccountName
	StakeCPU int64 `json:"stake_cpu"`
	StakeNet int64 `json:"stake_net"`
	Transfer bool
}

func (op *OpDelegateBW) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	return append(out, system.NewDelegateBW(op.From, op.To, eos.NewEOSAsset(op.StakeCPU), eos.NewEOSAsset(op.StakeNet), op.Transfer)), nil
}
