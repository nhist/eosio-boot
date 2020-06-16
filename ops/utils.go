package boot

import (
	"fmt"
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

// AN is a shortcut to create an AccountName
var AN = eos.AN

// PN is a shortcut to create a PermissionName
var PN = eos.PN


func decodeOpPublicKey(b *eosboot.Boot, opPubKey string) (ecc.PublicKey, error) {
	privateKey, err := b.GetBootseqKey(opPubKey)
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	pubKey, err := ecc.NewPublicKey(opPubKey)
	if err != nil {
		return ecc.PublicKey{}, fmt.Errorf("reading pubkey: %s", err)
	}
	return pubKey, nil
}
