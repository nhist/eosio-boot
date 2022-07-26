package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"go.uber.org/zap"
)

func init() {
	Register("system.setprods", &OpSetProds{})
}

type OpSetProds struct {
	Prods []producerKeyString
}

func (op *OpSetProds) RequireValidation() bool {
	return true
}

func (op *OpSetProds) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	var prodKeys []system.ProducerKey

	for _, key := range op.Prods {
		prodKey := system.ProducerKey{
			ProducerName: key.ProducerName,
		}
		pubKey, err := decodeOpPublicKey(c, key.BlockSigningKeyString)
		if err != nil {
			return err
		}
		prodKey.BlockSigningKey = pubKey
		prodKeys = append(prodKeys, prodKey)
	}

	pubKey, err := getBootKey(c)
	if err != nil {
		return err
	}

	if len(prodKeys) == 0 {
		prodKeys = []system.ProducerKey{system.ProducerKey{
			ProducerName:    AN("zswhq"),
			BlockSigningKey: pubKey,
		}}
	}

	var producers []string
	for _, key := range prodKeys {
		producers = append(producers, string(key.ProducerName))
	}
	c.Logger.Info("producers are set", zap.Strings("procuders", producers))

	in <- (*TransactionAction)(system.NewSetProds(prodKeys))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

type producerKeyString struct {
	ProducerName          eos.AccountName `json:"producer_name"`
	BlockSigningKeyString string          `json:"block_signing_key"`
}
