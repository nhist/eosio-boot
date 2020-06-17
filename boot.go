package boot

import (
	"context"
	"fmt"
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/dfuse-io/eosio-boot/content"
	"github.com/dfuse-io/eosio-boot/ops"
	"github.com/dfuse-io/eosio-boot/snapshot"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
	"os"
	"time"
)

type option func(b *Boot) *Boot

func WithMaxActionCountPerTrx(max int) option {
	return func(b *Boot) *Boot {
		b.maxActionCountPerTrx = max
		return b
	}
}

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
	maxActionCountPerTrx int
	Snapshot           snapshot.Snapshot
	WriteActions       bool
	HackVotingAccounts bool
}

func New(bootSequencePath string, targetAPI *eos.API, cachePath string, opts ...option) (b *Boot, err error) {
	b = &Boot{
		targetNetAPI:     targetAPI,
		bootSequencePath: bootSequencePath,
		contentManager:   content.NewManager(cachePath),
		maxActionCountPerTrx: 500,
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



	actions := make(chan interface{}, 500)
	go func() {
		defer close(actions)
		for _, step := range b.bootSequence.BootSequence {

			zlog.Info("action",
				zap.String("label", step.Label),
				zap.String("op", step.Op),
				zap.String("signer", step.Signer),
				zap.Bool("validate", step.Validate),
			)
			pubkey, err := b.getOpPubkey(step)
			if err != nil {
				zlog.Error("unable to get public key for operation", zap.Error(err))
				return
			}

			err = step.Data.Actions(pubkey, opConfig , actions)
			if err != nil {
				zlog.Error("unable to get actions for step", zap.String("ops", step.Op), zap.Error(err))
				return
			}

		}
	}()

	for {
		trxBundle := b.chunkifyActionChan(actions)
		if trxBundle == nil {
			// chunkify exited without given any chunks, channel must be closed
			break
		}

		b.targetNetAPI.SetCustomGetRequiredKeys(func(ctx context.Context, tx *eos.Transaction) (out []ecc.PublicKey, err error) {
			out = append(out, trxBundle.signer)
			return out, nil
		})

		err := Retry(25, time.Second, func() error {
			_, err := b.targetNetAPI.SignPushActions(ctx, trxBundle.actions...)
			if err != nil {
				zlog.Error("error pushing transaction bundle", zap.Error(err))
				return fmt.Errorf("push actions of transaciton bundle: %w", err)
			}
			return nil
		})
		if err != nil {
			zlog.Error("failed to push transaction bundle", zap.Error(err))
			return "", err
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


type transactionBundle struct {
	actions []*eos.Action
	signer ecc.PublicKey
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

func (b *Boot) chunkifyActionChan(actions chan interface{}) (*transactionBundle) {
	out := &transactionBundle{
		actions: []*eos.Action{},
	}
	for {
		if len(out.actions) > b.maxActionCountPerTrx {
			return out
		}

		act, ok := <- actions
		if !ok {
			// channel is closed, there is not transaction to process
			return nil
		}
		switch act.(type) {
		case *ops.TransactionBoundary:
			trxBoundary := act.(*ops.TransactionBoundary)
			out.signer = trxBoundary.Signer
			return out
		case *eos.Action:
			action := act.(*eos.Action)
			if action != nil {
				out.actions = append(out.actions, action)
			}
		default:
			panic(fmt.Sprintf("chunkify: unexpected type in action chan"))
		}
	}
	return nil
}


func (b *Boot) getOpPubkey(op *ops.OperationType)  (ecc.PublicKey, error){
	if op.Signer != "" {
		if privKey, found := b.bootseqKeys[op.Signer]; found {
			return privKey.PublicKey(), nil
		}
		return ecc.PublicKey{}, fmt.Errorf("cannot find private key in boot sequence with label %q", op.Signer)
	}

	pubKey, err := b.getBootKey()
	if err != nil {
		return ecc.PublicKey{}, err
	}
	return pubKey, nil
}

