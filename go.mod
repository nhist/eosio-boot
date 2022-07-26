module github.com/dfuse-io/eosio-boot

go 1.14

require (
	github.com/abourget/llerrgroup v0.2.0
	github.com/bronze1man/go-yaml2json v0.0.0-20150129175009-f6f64b738964
	github.com/eoscanada/eos-go v0.9.1-0.20200723180508-f68c7571db82
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	gopkg.in/olivere/elastic.v3 v3.0.75
)

replace github.com/eoscanada/eos-go => github.com/nhist/zswchain-go v3.0.0