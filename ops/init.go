package ops

import (
	"fmt"

	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.init", &OpSystemInit{})
}

type OpSystemInit struct {
	Version eos.Varuint32 `json:"version"`
	Core    string        `json:"core"`
}

func (op *OpSystemInit) RequireValidation() bool {
	return true
}

func (op *OpSystemInit) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	core, err := eos.StringToSymbol(op.Core)
	if err != nil {
		return fmt.Errorf("unable to convert system.init core %q to symbol: %w", op.Core, err)
	}
	in <- (*TransactionAction)(system.NewInitSystem(op.Version, core))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
