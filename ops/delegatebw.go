package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.delegate_bw", &OpDelegateBW{})
}


type OpDelegateBW struct {
	From     eos.AccountName
	To       eos.AccountName
	StakeCPU int64 `json:"stake_cpu"`
	StakeNet int64 `json:"stake_net"`
	Transfer bool
}

func (op *OpDelegateBW) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	return append(out, system.NewDelegateBW(op.From, op.To, eos.NewEOSAsset(op.StakeCPU), eos.NewEOSAsset(op.StakeNet), op.Transfer)), nil
}
