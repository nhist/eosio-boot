package boot

import (
	"context"

	"github.com/eoscanada/eos-go"
)

func (b *Boot) getProducerProtocolFeatures(ctx context.Context) ([]eos.ProtocolFeature, error) {
	return b.targetNetAPI.GetProducerProtocolFeatures(ctx)
}
