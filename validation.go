package boot

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dfuse-io/eosio-boot/config"
	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"
	"os"
	"time"
)

func (b *Boot) RunChainValidation(opConfig *config.OpConfig) (bool, error) {
	bootSeqMap := ActionMap{}
	bootSeq := []*eos.Action{}

	for _, step := range b.bootSequence.BootSequence {
		acts, err := step.Data.Actions(opConfig)
		if err != nil {
			return false, fmt.Errorf("validating: getting actions for step %q: %s", step.Op, err)
		}

		for _, stepAction := range acts {
			if stepAction == nil {
				continue
			}

			stepAction.SetToServer(true)
			data, err := eos.MarshalBinary(stepAction)
			if err != nil {
				return false, fmt.Errorf("validating: binary marshalling: %s", err)
			}
			key := sha2(data)

			// if _, ok := bootSeqMap[key]; ok {
			// 	// TODO: don't fatal here plz :)
			// 	log.Fatalf("Same action detected twice [%s] with key [%s]\n", stepAction.Name, key)
			// }
			bootSeqMap[key] = stepAction
			bootSeq = append(bootSeq, stepAction)
		}

	}

	err := b.validateTargetNetwork(bootSeqMap, bootSeq)
	if err != nil {
		zlog.Info("BOOT SEQUENCE VALIDATION FAILED:", zap.Error(err))
		return false, nil
	}

	zlog.Info("")
	zlog.Info("All good! Chain validation succeeded!")
	zlog.Info("")

	return true, nil
}

func (b *Boot) validateTargetNetwork(bootSeqMap ActionMap, bootSeq []*eos.Action) (err error) {
	expectedActionCount := len(bootSeq)
	validationErrors := make([]error, 0)

	b.pingTargetNetwork()

	// TODO: wait for target network to be up, and responding...
	zlog.Info("Pulling blocks from chain until we gathered all actions to validate:")
	blockHeight := 1
	actionsRead := 0
	seenMap := map[string]bool{}
	gotSomeTx := false
	backOff := false
	timeBetweenFetch := time.Duration(0)
	var timeLastNotFound time.Time

	for {
		time.Sleep(timeBetweenFetch)

		m, err := b.targetNetAPI.GetBlockByNum(context.Background(), uint32(blockHeight))
		if err != nil {
			if gotSomeTx && !backOff {
				backOff = true
				timeBetweenFetch = 500 * time.Millisecond
				timeLastNotFound = time.Now()

				time.Sleep(2000 * time.Millisecond)

				continue
			}

			zlog.Warn("Failed getting block num from target api", zap.String("message", err.Error()))
			time.Sleep(1 * time.Second)
			continue
		}

		blockHeight++

		zlog.Info("Receiving block", zap.Uint32("block_num", m.BlockNumber()), zap.String("producer", string(m.Producer)), zap.Int("trx_count", len(m.Transactions)))

		if !gotSomeTx && len(m.Transactions) > 2 {
			gotSomeTx = true
		}

		if !timeLastNotFound.IsZero() && timeLastNotFound.Before(time.Now().Add(-10*time.Second)) {
			b.flushMissingActions(seenMap, bootSeq)
		}

		for _, receipt := range m.Transactions {
			unpacked, err := receipt.Transaction.Packed.Unpack()
			if err != nil {
				zlog.Warn("Unable to unpack transaction, won't be able to fully validate", zap.Error(err))
				return fmt.Errorf("unpack transaction failed")
			}

			for _, act := range unpacked.Actions {
				act.SetToServer(false)
				data, err := eos.MarshalBinary(act)
				if err != nil {
					zlog.Error("Error marshalling an action", zap.Error(err))
					validationErrors = append(validationErrors, ValidationError{
						Err:               err,
						BlockNumber:       1, // extract from the block transactionmroot
						PackedTransaction: receipt.Transaction.Packed,
						Action:            act,
						RawAction:         data,
						ActionHexData:     hex.EncodeToString(act.HexData),
						Index:             actionsRead,
					})
					return err
				}
				key := sha2(data) // TODO: compute a hash here..

				if _, ok := bootSeqMap[key]; !ok {
					validationErrors = append(validationErrors, ValidationError{
						Err:               errors.New("not found"),
						BlockNumber:       1, // extract from the block transactionmroot
						PackedTransaction: receipt.Transaction.Packed,
						Action:            act,
						RawAction:         data,
						ActionHexData:     hex.EncodeToString(act.HexData),
						Index:             actionsRead,
					})
					zlog.Warn("INVALID action", zap.Int("action_read", actionsRead+1), zap.Int("expected_action_count", expectedActionCount), zap.String("account", string(act.Account)), zap.String("action", string(act.Name)))
				} else {
					seenMap[key] = true
					zlog.Info("validated action", zap.Int("action_read", actionsRead+1), zap.Int("expected_action_count", expectedActionCount), zap.String("account", string(act.Account)), zap.String("action", string(act.Name)))
				}

				actionsRead++
			}
		}

		if actionsRead == len(bootSeq) {
			break
		}

	}

	if len(validationErrors) > 0 {
		return ValidationErrors{Errors: validationErrors}
	}

	return nil
}

func (b *Boot) flushMissingActions(seenMap map[string]bool, bootSeq []*eos.Action) {
	fl, err := os.Create("missing_actions.jsonl")
	if err != nil {
		zlog.Error("Couldn't write to `missing_actions.jsonl`:", zap.Error(err))
		return
	}
	defer fl.Close()

	// TODO: print all actions that are still MISSING to `missing_actions.jsonl`.
	zlog.Info("Flushing unseen transactions to `missing_actions.jsonl` up until this point.")

	for _, act := range bootSeq {
		act.SetToServer(true)
		data, _ := eos.MarshalBinary(act)
		key := sha2(data)

		if !seenMap[key] {
			act.SetToServer(false)
			data, _ := json.Marshal(act)
			fl.Write(data)
			fl.Write([]byte("\n"))
		}
	}
}