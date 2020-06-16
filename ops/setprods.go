package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"go.uber.org/zap"
)

func init() {
	Register("system.setprods", &OpSetProds{})
}

type OpSetProds struct {
	Prods []producerKeyString
}

func (op *OpSetProds) Actions(c *config.OpConfig) (out []*eos.Action, err error) {
	var prodKeys []system.ProducerKey

	for _, key := range op.Prods {
		prodKey := system.ProducerKey{
			ProducerName: key.ProducerName,
		}
		pubKey, err := decodeOpPublicKey(c, key.BlockSigningKeyString)
		if err != nil {
			return nil, err
		}
		prodKey.BlockSigningKey = pubKey
		prodKeys = append(prodKeys, prodKey)
	}

	pubKey, err := getBootKey(c)
	if err != nil {
		return nil, err
	}

	if len(prodKeys) == 0 {
		prodKeys = []system.ProducerKey{system.ProducerKey{
			ProducerName:    AN("eosio"),
			BlockSigningKey: pubKey,
		}}
	}

	var producers []string
	for _, key := range prodKeys {
		producers = append(producers, string(key.ProducerName))
	}
	zlog.Info("producers are set", zap.Strings("procuders", producers))

	out = append(out, system.NewSetProds(prodKeys))
	return
}

type producerKeyString struct {
	ProducerName          eos.AccountName `json:"producer_name"`
	BlockSigningKeyString string          `json:"block_signing_key"`
}


