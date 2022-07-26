package ops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_encodePayload(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		expectData  interface{}
		expectError bool
	}{
		{
			name:        "sunny path",
			payload:     "[\"0\", \"4,EOS\"]",
			expectData:  []interface{}{"0", "4,EOS"},
			expectError: false,
		},
		{
			name:        "sunny path",
			payload:     "{\"from\" : \"eosio\", \"to\" : \"eosio3\", \"quantity\": \"10 EOS\", \"memo\": \"custom trx\"}",
			expectData:  map[string]interface{}{"from": "zswhq", "memo": "custom trx", "quantity": "10 EOS", "to": "eosio3"},
			expectError: false,
		},
		{
			name:        "sunny path",
			payload:     "asdfaf",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := encodePayload(test.payload)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectData, data)
			}
		})
	}

}
