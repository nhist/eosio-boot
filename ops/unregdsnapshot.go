package boot

import (
	"fmt"
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/dfuse-io/eosio-boot/unregd"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
	"go.uber.org/zap"
)

func init() {
	eosboot.Register("snapshot.load_unregistered", &OpInjectUnregdSnapshot{})
}

type OpInjectUnregdSnapshot struct {
	TestnetTruncateSnapshot int `json:"TESTNET_TRUNCATE_SNAPSHOT"`
}

func (op *OpInjectUnregdSnapshot) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	snapshotFile, err := b.GetContentsCacheRef("snapshot_unregistered.csv")
	if err != nil {
		return nil, err
	}

	rawSnapshot, err := b.ReadFromCache(snapshotFile)
	if err != nil {
		return nil, fmt.Errorf("reading snapshot file: %s", err)
	}

	snapshotData, err := eosboot.NewUnregdSnapshot(rawSnapshot)
	if err != nil {
		return nil, fmt.Errorf("loading snapshot csv: %s", err)
	}

	if len(snapshotData) == 0 {
		return nil, fmt.Errorf("snapshot is empty or not loaded")
	}

	for idx, hodler := range snapshotData {
		if trunc := op.TestnetTruncateSnapshot; trunc != 0 {
			if idx == trunc {
				zlog.Debug("- DEBUG: truncated unreg'd snapshot", zap.Int("row", trunc))
				break
			}
		}

		//system.NewDelegatedNewAccount(AN("eosio"), AN(hodler.AccountName), AN("eosio.unregd"))

		out = append(out,
			unregd.NewAdd(hodler.EthereumAddress, hodler.Balance),
			token.NewTransfer(AN("eosio"), AN("eosio.unregd"), hodler.Balance, "Future claim"),
			nil,
		)
	}

	return
}