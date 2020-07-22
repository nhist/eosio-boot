package config

import (
	"context"
	"fmt"

	"github.com/dfuse-io/eosio-boot/content"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
)

type abiCache struct {
	nodeApi *eos.API
	abis    map[eos.AccountName]*eos.ABI
}

func newAbiCache(nodeApi *eos.API) *abiCache {
	return &abiCache{
		nodeApi: nodeApi,
		abis:    map[eos.AccountName]*eos.ABI{},
	}
}

func (a *abiCache) SetABI(accountName eos.AccountName, abi *eos.ABI) {
	a.abis[accountName] = abi
}

func (a *abiCache) GetABI(accountName eos.AccountName) (*eos.ABI, error) {
	if abi, found := a.abis[accountName]; found {
		return abi, nil
	}

	resp, err := a.nodeApi.GetABI(context.Background(), accountName)
	if err != nil {
		return nil, fmt.Errorf("ABI not found in cache and could not retrieve from chain: %w", err)
	}

	abi := &resp.ABI
	a.SetABI(accountName, abi)

	return abi, nil
}

type OpConfig struct {
	contentRefs    []*content.ContentRef
	privateKeys    map[string]*ecc.PrivateKey
	contentManager *content.Manager
	AbiCache       *abiCache
	Logger         *zap.Logger
}

func NewOpConfig(contentRefs []*content.ContentRef, contentManager *content.Manager, privateKeys map[string]*ecc.PrivateKey, api *eos.API) *OpConfig {
	return &OpConfig{
		contentRefs:    contentRefs,
		privateKeys:    privateKeys,
		contentManager: contentManager,
		AbiCache:       newAbiCache(api),
		Logger:         zap.NewNop(),
	}
}

func (c OpConfig) HackVotingAccounts() bool {
	return false
}

func (c OpConfig) ReadFromCache(ref string) ([]byte, error) {
	return c.contentManager.ReadFromCache(ref)

}

func (c OpConfig) GetContentsCacheRef(filename string) (string, error) {
	for _, fl := range c.contentRefs {
		if fl.Name == filename {
			return fl.URL, nil
		}
	}
	return "", fmt.Errorf("%q not found in target contents", filename)
}

func (c OpConfig) GetPrivateKey(label string) (*ecc.PrivateKey, error) {
	if _, found := c.privateKeys[label]; found {
		return c.privateKeys[label], nil
	}
	return nil, fmt.Errorf("bootseq does not contain key with label %q", label)

}

func (c OpConfig) FileNameFromCache(ref string) string {
	return c.contentManager.FileNameFromCache(ref)
}
