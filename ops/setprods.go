package boot

import (
	"fmt"
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"go.uber.org/zap"
)

func init() {
	eosboot.Register("system.setprods", &OpSetProds{})
}

type OpSetProds struct {
	Prods []producerKeyString
}

func (op *OpSetProds) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	var prodKeys []system.ProducerKey

	for _, key := range op.Prods {
		prodKey := system.ProducerKey{
			ProducerName: key.ProducerName,
		}
		pubKey, err := decodeOpPublicKey(b, key.BlockSigningKeyString)
		if err != nil {
			return nil, err
		}
		prodKey.BlockSigningKey = pubKey
		prodKeys = append(prodKeys, prodKey)
	}

	pubKey, err := getBootKey(b)
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


// this is use to support ephemeral key
func getBootKey(b *eosboot.Boot) (ecc.PublicKey, error) {
	privateKey, err := b.GetBootseqKey("boot")
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	privateKey, err = b.GetBootseqKey("ephemeral")
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	return ecc.PublicKey{}, fmt.Errorf("cannot find boot/ephemeral key")
}
