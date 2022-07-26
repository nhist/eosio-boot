package ops

import (
	"fmt"

	"github.com/dfuse-io/eosio-boot/config"
	"github.com/dfuse-io/eosio-boot/snapshot"
	"github.com/dfuse-io/eosio-boot/unregd"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"go.uber.org/zap"
)

func init() {
	Register("snapshot.load_unregistered", &OpInjectUnregdSnapshot{})
}

type OpInjectUnregdSnapshot struct {
	TestnetTruncateSnapshot int `json:"TESTNET_TRUNCATE_SNAPSHOT"`
}

func (op *OpInjectUnregdSnapshot) RequireValidation() bool {
	return true
}

func (op *OpInjectUnregdSnapshot) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	snapshotFile, err := c.GetContentsCacheRef("snapshot_unregistered.csv")
	if err != nil {
		return err
	}

	rawSnapshot, err := c.ReadFromCache(snapshotFile)
	if err != nil {
		return fmt.Errorf("reading snapshot file: %s", err)
	}

	snapshotData, err := snapshot.NewUnregd(rawSnapshot)
	if err != nil {
		return fmt.Errorf("loading snapshot csv: %s", err)
	}

	if len(snapshotData) == 0 {
		return fmt.Errorf("snapshot is empty or not loaded")
	}

	for idx, hodler := range snapshotData {
		if trunc := op.TestnetTruncateSnapshot; trunc != 0 {
			if idx == trunc {
				c.Logger.Debug("- DEBUG: truncated unreg'd snapshot", zap.Int("row", trunc))
				break
			}
		}

		//system.NewDelegatedNewAccount(AN("zswhq"), AN(hodler.AccountName), AN("eosio.unregd"))
		in <- (*TransactionAction)(unregd.NewAdd(hodler.EthereumAddress, hodler.Balance))
		in <- (*TransactionAction)(token.NewTransfer(AN("zswhq"), AN("eosio.unregd"), hodler.Balance, "Future claim"))
		in <- EndTransaction(opPubkey) // End Transaction
	}

	return nil
}
