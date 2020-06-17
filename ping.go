package boot

import (
	"context"
	"go.uber.org/zap"
	"time"
)

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