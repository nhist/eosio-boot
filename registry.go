package boot

import (
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	_ "github.com/dfuse-io/eosio-boot/ops"
	"go.uber.org/zap"
	"reflect"
)

type Operation interface {
	Actions(b *Boot) ([]*eos.Action, error)
}

var operationsRegistry = map[string]Operation{}

func Register(key string, operation Operation) {
	if key == "" {
		zlog.Fatal("key cannot be blank")
	} else if _, ok := operationsRegistry[key]; ok {
		zlog.Fatal("already registered", zap.String("key", key))
	}
	operationsRegistry[key] = operation

}

type OperationType struct {
	Op     string
	Signer string
	Label  string
	Data   Operation
}

func (o *OperationType) UnmarshalJSON(data []byte) error {
	opData := struct {
		Op     string
		Signer string
		Label  string
		Data   json.RawMessage
	}{}
	if err := json.Unmarshal(data, &opData); err != nil {
		return err
	}

	opType, found := operationsRegistry[opData.Op]
	if !found {
		return fmt.Errorf("operation type %q invalid, use one of: %q", opData.Op, operationsRegistry)
	}

	objType := reflect.TypeOf(opType).Elem()
	obj := reflect.New(objType).Interface()

	if len(opData.Data) != 0 {
		err := json.Unmarshal(opData.Data, &obj)
		if err != nil {
			return fmt.Errorf("operation type %q invalid, error decoding: %s", opData.Op, err)
		}
	}

	opIface, ok := obj.(Operation)
	if !ok {
		return fmt.Errorf("operation type %q isn't an op", opData.Op)
	}

	*o = OperationType{
		Op:     opData.Op,
		Label:  opData.Label,
		Signer: opData.Signer,
		Data:   opIface,
	}

	return nil
}
