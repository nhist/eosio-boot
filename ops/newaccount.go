package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.newaccount", &OpNewAccount{})
}


type OpNewAccount struct {
	Creator    eos.AccountName
	NewAccount eos.AccountName `json:"new_account"`
	Pubkey     string
	RamBytes   uint32 `json:"ram_bytes"`
}

func (op *OpNewAccount) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	pubKey, err := decodeOpPublicKey(c, op.Pubkey)
	if err != nil {
		return err
	}

	in <- system.NewNewAccount(op.Creator, op.NewAccount, pubKey)

	if op.RamBytes > 0 {
		in <- system.NewBuyRAMBytes(op.Creator, op.NewAccount, op.RamBytes)
	}
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
