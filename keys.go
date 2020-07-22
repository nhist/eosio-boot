package boot

import (
	"context"
	"fmt"
	"strings"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
)

func (b *Boot) setKeys() error {
	if b.keyBag == nil {
		b.logger.Info("key bag not preset")
		b.keyBag = eos.NewKeyBag()
	}

	for label, privKey := range b.bootseqKeys {
		privKeyStr := privKey.String()
		b.logger.Info("adding bootseq key to keybag",
			zap.String("key_tag", label),
			zap.String("pub_key", privKey.PublicKey().String()),
			zap.String("priv_key_prefix", privKey.String()[:4]),
			zap.String("priv_key", privKey.String()[len(privKey.String())-4:]),
		)
		b.keyBag.Add(privKeyStr)
	}

	return nil
}

func (b *Boot) attachKeysOnTargetNode(ctx context.Context) error {
	// Store keys in wallet, to sign `SetCode` and friends..
	b.targetNetAPI.SetSigner(b.keyBag)
	return nil
}

func (b *Boot) parseBootseqKeys() error {
	for label, key := range b.bootSequence.Keys {
		privKey, err := ecc.NewPrivateKey(strings.TrimSpace(key))
		if err != nil {
			return fmt.Errorf("unable to correctly decode %q private key %q: %s", label, key, err)
		}
		b.bootseqKeys[label] = privKey
	}
	return nil
}


func (b *Boot) getBootKey() (ecc.PublicKey, error) {
	privKey, err := b.getBootseqKey("boot")
	if err == nil {
		return privKey.PublicKey(), nil
	}

	privKey, err = b.getBootseqKey("ephemeral")
	if err == nil {
		return privKey.PublicKey(), nil
	}

	return ecc.PublicKey{}, fmt.Errorf("unable to find boot or ephemeral key in boot seq")

}