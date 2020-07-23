package ops

import (
	"encoding/json"
	"fmt"

	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

func init() {
	Register("system.pushtransaction", &OpPushTransaction{})
}

type OpPushTransaction struct {
	Contract   eos.AccountName
	Action     eos.ActionName
	Actor      eos.AccountName
	Permission eos.PermissionName
	Payload    map[string]interface{}
}

func (op *OpPushTransaction) RequireValidation() bool {
	return true
}

func (op *OpPushTransaction) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	cnt, err := json.Marshal(op.Payload)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	abi, err := c.AbiCache.GetABI(op.Contract)
	if err != nil {
		return fmt.Errorf("cannot retrieve ABI for account %q encode payload: %w", op.Contract, err)
	}

	actionBinary, err := abi.EncodeAction(op.Action, []byte(cnt))

	action := &eos.Action{
		Account: op.Contract,
		Name:    op.Action,
		Authorization: []eos.PermissionLevel{
			{Actor: op.Actor, Permission: op.Permission},
		},
		ActionData: eos.NewActionDataFromHexData(actionBinary),
	}
	in <- (*TransactionAction)(action)
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func encodePayload(payload string) (interface{}, error) {
	var hashData map[string]interface{}
	err := json.Unmarshal([]byte(payload), &hashData)
	if err == nil {
		return hashData, nil
	}

	var data []interface{}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return nil, fmt.Errorf("unsupported payload format: %w", err)
	}
	return data, nil
}
