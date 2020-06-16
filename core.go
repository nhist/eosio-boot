package boot

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/dfuse-io/eosio-boot/content"
	"github.com/dfuse-io/eosio-boot/snapshot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
	"os"
	"time"
)

type option func(b *Boot) *Boot

func WithKeyBag(keyBag *eos.KeyBag) option {
	return func(b *Boot) *Boot {
		b.keyBag = keyBag
		return b
	}
}

type Boot struct {
	bootSequencePath     string
	targetNetAPI         *eos.API
	bootstrappingEnabled bool
	genesisPath          string
	bootSequence         *BootSeq

	contentManager *content.Manager
	keyBag      *eos.KeyBag
	bootseqKeys map[string]*ecc.PrivateKey

	Snapshot           snapshot.Snapshot
	WriteActions       bool
	HackVotingAccounts bool
}

func New(bootSequencePath string, targetAPI *eos.API, cachePath string, opts ...option) (b *Boot, err error) {
	b = &Boot{
		targetNetAPI:     targetAPI,
		bootSequencePath: bootSequencePath,
		contentManager:   content.NewManager(cachePath),
		bootseqKeys:      map[string]*ecc.PrivateKey{},
	}
	for _, opt := range opts {
		b = opt(b)
	}

	b.bootSequence, err = readBootSeq(b.bootSequencePath)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Boot) Revision() string {
	return b.bootSequence.Checksum
}

func (b *Boot) getBootseqKey(label string) (*ecc.PrivateKey, error) {
	if _, found := b.bootseqKeys[label]; found {
		return b.bootseqKeys[label], nil
	}
	return nil, fmt.Errorf("bootseq does not contain key with label %q", label)
}

func (b *Boot) Run() (checksums string, err error) {
	ctx := context.Background()

	zlog.Debug("parsing boot sequence keys")
	if err := b.parseBootseqKeys(); err != nil {
		return "", err
	}

	zlog.Debug("downloading references")
	if err := b.contentManager.Download(b.bootSequence.Contents); err != nil {
		return "", err
	}

	zlog.Debug("setting boot keys")
	if err := b.setKeys(); err != nil {
		return "", err
	}

	if err := b.attachKeysOnTargetNode(ctx); err != nil {
		return "", err
	}

	b.pingTargetNetwork()

	opConfig := config.NewOpConfig(
		b.bootSequence.Contents,
		b.contentManager,
		b.bootseqKeys,
		)

	//eos.Debug = true
	for _, step := range b.bootSequence.BootSequence {
		zlog.Info("action", zap.String("label", step.Label), zap.String("op", step.Op))

		acts, err := step.Data.Actions(opConfig)
		if err != nil {
			return "", fmt.Errorf("getting actions for step %q: %s", step.Op, err)
		}

		if step.Signer != "" {
			zlog.Info("setting required keys", zap.String("signer", step.Signer))
			b.targetNetAPI.SetCustomGetRequiredKeys(func(ctx context.Context, tx *eos.Transaction) (out []ecc.PublicKey, err error) {
				privKey, err := b.getBootseqKey(step.Signer)
				if err != nil {
					return nil, err
				}
				out = append(out, privKey.PublicKey())
				return out, nil
			})
		} else {
			zlog.Info("setting required key to boot/ephemeral key")
			b.targetNetAPI.SetCustomGetRequiredKeys(func(ctx context.Context, tx *eos.Transaction) (out []ecc.PublicKey, err error) {
				privKey, err := b.getBootseqKey("boot")
				if err == nil {
					out = append(out, privKey.PublicKey())
					return out, nil
				}

				privKey, err = b.getBootseqKey("ephemeral")
				if err == nil {
					out = append(out, privKey.PublicKey())
					return out, nil
				}

				return nil, fmt.Errorf("unable to find boot or ephemeral key in boot seq")
			})
		}

		if len(acts) != 0 {
			for idx, chunk := range ChunkifyActions(acts) {
				for _, c := range chunk {
					zlog.Info("processing chunk", zap.String("action", string(c.Name)))
				}
				err := Retry(25, time.Second, func() error {

					_, err := b.targetNetAPI.SignPushActions(ctx, chunk...)
					if err != nil {
						zlog.Error("error pushing transaction", zap.String("op", step.Op), zap.Int("idx", idx), zap.Error(err))
						return fmt.Errorf("push actions for step %q, chunk %d: %s", step.Op, idx, err)
					}
					return nil
				})
				if err != nil {
					zlog.Info(" failed")
					return "", err
				}
			}
		}
	}

	zlog.Info("Waiting 2 seconds for transactions to flush to blocks")
	time.Sleep(2 * time.Second)

	// FIXME: don't do chain validation here..
	isValid, err := b.RunChainValidation(opConfig)
	if err != nil {
		return "", fmt.Errorf("chain validation: %s", err)
	}
	if !isValid {
		zlog.Info("WARNING: chain invalid, destroying network if possible")
		os.Exit(0)
	}

	return b.bootSequence.Checksum, nil
}




type ActionMap map[string]*eos.Action

type ValidationError struct {
	Err               error
	BlockNumber       int
	Action            *eos.Action
	RawAction         []byte
	Index             int
	ActionHexData     string
	PackedTransaction *eos.PackedTransaction
}

func (e ValidationError) Error() string {
	s := fmt.Sprintf("Action [%d][%s::%s] absent from blocks\n", e.Index, e.Action.Account, e.Action.Name)

	data, err := json.Marshal(e.Action)
	if err != nil {
		s += fmt.Sprintf("    json generation err: %s\n", err)
	} else {
		s += fmt.Sprintf("    json data: %s\n", string(data))
	}
	s += fmt.Sprintf("    hex data: %s\n", hex.EncodeToString(e.RawAction))
	s += fmt.Sprintf("    error: %s\n", e.Err.Error())

	return s
}

type ValidationErrors struct {
	Errors []error
}

func (v ValidationErrors) Error() string {
	s := ""
	for _, err := range v.Errors {
		s += ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n"
		s += err.Error()
		s += "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n"
	}

	return s
}

func (b *Boot) pingTargetNetwork() {
	zlog.Info("Pinging target node at ", zap.String("url", b.targetNetAPI.BaseURL))
	for {
		info, err := b.targetNetAPI.GetInfo(context.Background())
		if err != nil {
			zlog.Warn("target node", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		if info.HeadBlockNum < 2 {
			zlog.Info("target node: still no blocks in")
			zlog.Info(".")
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

	zlog.Info(" touchdown!")
}



func ChunkifyActions(actions []*eos.Action) (out [][]*eos.Action) {
	currentChunk := []*eos.Action{}
	for _, act := range actions {
		if act == nil {
			if len(currentChunk) != 0 {
				out = append(out, currentChunk)
			}
			currentChunk = []*eos.Action{}
		} else {
			currentChunk = append(currentChunk, act)
		}
	}
	if len(currentChunk) > 0 {
		out = append(out, currentChunk)
	}
	return
}