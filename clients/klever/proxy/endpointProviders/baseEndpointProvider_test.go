package endpointProviders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseEndpointProvider(t *testing.T) {
	t.Parallel()

	base := &baseEndpointProvider{}
	assert.Equal(t, networkConfig, base.GetNetworkConfig())
	assert.Equal(t, networkEconomics, base.GetNetworkEconomics())
	assert.Equal(t, enableEpochsConfig, base.GetEnableEpochsConfig())
	assert.Equal(t, "address/addressAsBech32", base.GetAccount("addressAsBech32"))
	assert.Equal(t, estimateTransactionFees, base.GetEstimateTransactionFees())
	assert.Equal(t, sendTransaction, base.GetSendTransaction())
	assert.Equal(t, sendMultipleTransactions, base.GetSendMultipleTransactions())
	assert.Equal(t, "transaction/hex/status", base.GetTransactionStatus("hex"))
	assert.Equal(t, "transaction/hex", base.GetTransactionInfo("hex"))
	assert.Equal(t, vmValues, base.GetVmValues())
	assert.Equal(t, "address/erd1address/esdt/TKN-001122", base.GetESDTTokenData("erd1address", "TKN-001122"))
	assert.Equal(t, "address/erd1address/nft/TKN-001122/nonce/37", base.GetNFTTokenData("erd1address", "TKN-001122", 37))
}
