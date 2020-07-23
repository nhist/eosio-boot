package boot

import (
	"context"

	"github.com/eoscanada/eos-go"
)

func (b *Boot) getProducerProtocolFeatures(ctx context.Context) ([]eos.ProtocolFeature, error) {
	b.targetNetAPI.Debug = true
	defer func() {
		b.targetNetAPI.Debug = false
	}()
	return b.targetNetAPI.GetProducerProtocolFeatures(ctx)
}
