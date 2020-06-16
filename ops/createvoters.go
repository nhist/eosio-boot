package boot

import (
	"bytes"
	"fmt"
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"github.com/eoscanada/eos-go/token"
)

func init() {
	eosboot.Register("system.create_voters", &OpCreateVoters{})
}


type OpCreateVoters struct {
	Creator eos.AccountName
	Pubkey  string
	Count   int
}

func (op *OpCreateVoters) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	pubKey, err := decodeOpPublicKey(b, op.Pubkey)
	if err != nil {
		return nil, err
	}

	for i := 0; i < op.Count; i++ {
		voterName := eos.AccountName(voterName(i))
		fmt.Println("Creating voter: ", voterName)
		out = append(out, system.NewNewAccount(op.Creator, voterName, pubKey))
		out = append(out, token.NewTransfer(op.Creator, voterName, eos.NewEOSAsset(1000000000), ""))
		out = append(out, system.NewBuyRAMBytes(AN("eosio"), voterName, 8192)) // 8kb gift !
		out = append(out, system.NewDelegateBW(AN("eosio"), voterName, eos.NewEOSAsset(10000), eos.NewEOSAsset(10000), true))
	}
	return
}
const charset = "abcdefghijklmnopqrstuvwxyz"

func voterName(index int) string {
	padding := string(bytes.Repeat([]byte{charset[index]}, 7))
	return "voter" + padding
}
