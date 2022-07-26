package ops

import (
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	Register("system.resign_accounts", &OpResignAccounts{})
}

type OpResignAccounts struct {
	Accounts            []eos.AccountName
	TestnetKeepAccounts bool `json:"TESTNET_KEEP_ACCOUNTS"`
}

func (op *OpResignAccounts) RequireValidation() bool {
	return true
}

func (op *OpResignAccounts) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	if op.TestnetKeepAccounts {
		c.Logger.Debug("keeping system accounts around, for testing purposes.")
		return nil
	}

	systemAccount := AN("zswhq")
	prodsAccount := AN("eosio.prods") // this is a special system account that is granted by 2/3 + 1 of the current BP schedule.

	eosioPresent := false
	for _, acct := range op.Accounts {
		if acct == systemAccount {
			eosioPresent = true
			continue
		}

		in <- (*TransactionAction)(system.NewUpdateAuth(acct, PN("active"), PN("owner"), eos.Authority{
			Threshold: 1,
			Accounts: []eos.PermissionLevelWeight{
				eos.PermissionLevelWeight{
					Permission: eos.PermissionLevel{
						Actor:      AN("zswhq"),
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("active")))
		in <- (*TransactionAction)(system.NewUpdateAuth(acct, PN("owner"), PN(""), eos.Authority{
			Threshold: 1,
			Accounts: []eos.PermissionLevelWeight{
				eos.PermissionLevelWeight{
					Permission: eos.PermissionLevel{
						Actor:      AN("zswhq"),
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("owner")))

	}

	if eosioPresent {
		in <- (*TransactionAction)(system.NewUpdateAuth(systemAccount, PN("active"), PN("owner"), eos.Authority{
			Threshold: 1,
			Accounts: []eos.PermissionLevelWeight{
				eos.PermissionLevelWeight{
					Permission: eos.PermissionLevel{
						Actor:      prodsAccount,
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("active")))
		in <- (*TransactionAction)(system.NewUpdateAuth(systemAccount, PN("owner"), PN(""), eos.Authority{
			Threshold: 1,
			Accounts: []eos.PermissionLevelWeight{
				eos.PermissionLevelWeight{
					Permission: eos.PermissionLevel{
						Actor:      prodsAccount,
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("owner")))
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
