package ops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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

	codeAction, err := system.NewSetCode(op.Account, c.FileNameFromCache(wasmFileRef))
	if err != nil {
		return fmt.Errorf("NewSetCode %s: %s", op.ContractNameRef, err)
	}

	abi, err := retrieveABIfromRef(c.FileNameFromCache(abiFileRef))
	if err != nil {
		return fmt.Errorf("unable to read ABI %s: %s", abiFileRef, err)
	}

	abiAction, err := system.NewSetAbiFromAbi(op.Account, *abi)
	if err != nil {
		return fmt.Errorf("NewSetAbiFromAbi %s: %s", op.ContractNameRef, err)
	}

	c.AbiCache.SetABI(op.Account, abi)
	for _, act := range []*eos.Action{codeAction, abiAction} {
		in <- (*TransactionAction)(act)
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func retrieveABIfromRef(abiPath string) (*eos.ABI, error) {
	abiContent, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}
	if len(abiContent) == 0 {
		return nil, fmt.Errorf("unable to unmarshal abi with 0 bytes")
	}

	var abiDef eos.ABI
	if err := json.Unmarshal(abiContent, &abiDef); err != nil {
		return nil, fmt.Errorf("unmarshal ABI file: %s", err)
	}

	return &abiDef, nil
}
