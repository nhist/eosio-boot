package boot

import (
	"fmt"
	"github.com/eoscanada/eos-go"
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go/system"

)

func init() {
	eosboot.Register("system.setcode", &OpSetCode{})
}

type OpSetCode struct {
	Account         eos.AccountName
	ContractNameRef string `json:"contract_name_ref"`
}

func (op *OpSetCode) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	wasmFileRef, err := b.GetContentsCacheRef(fmt.Sprintf("%s.wasm", op.ContractNameRef))
	if err != nil {
		return nil, err
	}
	abiFileRef, err := b.GetContentsCacheRef(fmt.Sprintf("%s.abi", op.ContractNameRef))
	if err != nil {
		return nil, err
	}

	actions, err := system.NewSetContract(
		op.Account,
		b.FileNameFromCache(wasmFileRef),
		b.FileNameFromCache(abiFileRef),
	)
	if err != nil {
		return nil, fmt.Errorf("NewSetContract %s: %s", op.ContractNameRef, err)
	}

	return actions, nil
}


