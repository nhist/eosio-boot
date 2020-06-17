package ops

import (
	"fmt"
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.setcode", &OpSetCode{})
}

type OpSetCode struct {
	Account         eos.AccountName
	ContractNameRef string `json:"contract_name_ref"`
}


func (op *OpSetCode) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	wasmFileRef, err := c.GetContentsCacheRef(fmt.Sprintf("%s.wasm", op.ContractNameRef))
	if err != nil {
		return err
	}
	abiFileRef, err := c.GetContentsCacheRef(fmt.Sprintf("%s.abi", op.ContractNameRef))
	if err != nil {
		return err
	}

	actions, err := system.NewSetContract(
		op.Account,
		c.FileNameFromCache(wasmFileRef),
		c.FileNameFromCache(abiFileRef),
	)
	if err != nil {
		return fmt.Errorf("NewSetContract %s: %s", op.ContractNameRef, err)
	}

	for _, act := range actions {
		in <- act
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}


