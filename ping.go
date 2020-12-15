package boot

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (b *Boot) waitTargetNodeToBeUp() {
	b.logger.Info("Pinging target node at ", zap.String("url", b.targetNetAPI.BaseURL))
	for {
		info, err := b.targetNetAPI.GetInfo(context.Background())
		if err != nil {
			b.logger.Warn("target node", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		if info.HeadBlockNum < 2 {
			b.logger.Info("target node: still no blocks in")
			b.logger.Info(".")
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}
}
