package boot

import (
	eosboot "github.com/dfuse-io/eosio-boot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
)

func init() {
	eosboot.Register("system.resign_accounts", &OpResignAccounts{})
}


type OpResignAccounts struct {
	Accounts            []eos.AccountName
	TestnetKeepAccounts bool `json:"TESTNET_KEEP_ACCOUNTS"`
}

func (op *OpResignAccounts) Actions(b *eosboot.Boot) (out []*eos.Action, err error) {
	if op.TestnetKeepAccounts {
		zlog.Debug("keeping system accounts around, for testing purposes.")
		return
	}

	systemAccount := AN("eosio")
	prodsAccount := AN("eosio.prods") // this is a special system account that is granted by 2/3 + 1 of the current BP schedule.

	eosioPresent := false
	for _, acct := range op.Accounts {
		if acct == systemAccount {
			eosioPresent = true
			continue
		}

		out = append(out,
			system.NewUpdateAuth(acct, PN("active"), PN("owner"), eos.Authority{
				Threshold: 1,
				Accounts: []eos.PermissionLevelWeight{
					eos.PermissionLevelWeight{
						Permission: eos.PermissionLevel{
							Actor:      AN("eosio"),
							Permission: PN("active"),
						},
						Weight: 1,
					},
				},
			}, PN("active")),
			system.NewUpdateAuth(acct, PN("owner"), PN(""), eos.Authority{
				Threshold: 1,
				Accounts: []eos.PermissionLevelWeight{
					eos.PermissionLevelWeight{
						Permission: eos.PermissionLevel{
							Actor:      AN("eosio"),
							Permission: PN("active"),
						},
						Weight: 1,
					},
				},
			}, PN("owner")),
		)
	}

	if eosioPresent {
		out = append(out,
			system.NewUpdateAuth(systemAccount, PN("active"), PN("owner"), eos.Authority{
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
			}, PN("active")),
			system.NewUpdateAuth(systemAccount, PN("owner"), PN(""), eos.Authority{
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
			}, PN("owner")),
		)
	}

	out = append(out, nil)

	return
}