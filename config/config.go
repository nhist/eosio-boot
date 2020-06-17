package config

import (
	"fmt"
	"github.com/dfuse-io/eosio-boot/content"
	"github.com/eoscanada/eos-go/ecc"
)

type OpConfig struct {
	contentRefs []*content.ContentRef
	privateKeys map[string]*ecc.PrivateKey
	contentManager     *content.Manager
}

func NewOpConfig(contentRefs []*content.ContentRef, contentManager *content.Manager, privateKeys map[string]*ecc.PrivateKey) *OpConfig {
	return &OpConfig{
		contentRefs:    contentRefs,
		privateKeys:    privateKeys,
		contentManager: contentManager,
	}
}

func (c OpConfig) HackVotingAccounts() bool {
	return false
}

func (c OpConfig) ReadFromCache(ref string) ([]byte, error) {
	return c.contentManager.ReadFromCache(ref)

}

func (c OpConfig) GetContentsCacheRef(filename  string) (string, error) {
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