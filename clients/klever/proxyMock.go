package klever

import (
	"context"
	"encoding/hex"

	"github.com/klever-io/klv-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-sdk-go/data"
)

// TODO: check if safe contract uses token id to identify the token, or the contract address
// random token id
const tokenID = "tck-000001"

// TODO: remove CreateMockProxyKLV when real proxy is implemented
// CreateMockProxyKLV returns a interactors.ProxyStub, to be used as a klever client proxy and
// be used to run the bridge while real proxy implementation is in development
func CreateMockProxyKLV() *interactors.ProxyStub {
	return &interactors.ProxyStub{
		ExecuteVMQueryCalled:   executeVmQueryMock,
		GetNetworkStatusCalled: getNetworkStatusMock,
		GetNetworkConfigCalled: getNetworkConfigMock,
	}
}

// executeVmQueryMock is the function called when querying info from contracts
func executeVmQueryMock(_ context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	returningBytes := [][]byte{}
	switch vmRequest.FuncName {
	// getTokenIdForErc20AddressFuncName query the proxy (MultisigContractAddress) for a token id given a specific erc20 address
	case getTokenIdForErc20AddressFuncName:
		returningBytes = [][]byte{[]byte(tokenID)}

	// isMintBurnTokenFuncName checks if the token specified by the tokenID is MintBurnToken
	case isMintBurnTokenFuncName:
		if vmRequest.Args[0] == hex.EncodeToString([]byte(tokenID)) {
			// returns the byte response of a boolean
			byteResponseTrue := []byte{1}
			returningBytes = [][]byte{byteResponseTrue}
			break
		}
	}

	return &data.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnCode: okCodeAfterExecution,
			ReturnData: returningBytes,
		},
	}, nil
}

func getNetworkStatusMock(_ context.Context, _ uint32) (*data.NetworkStatus, error) {
	expectedNonce := uint64(0)

	return &data.NetworkStatus{
		Nonce: expectedNonce,
	}, nil
}

func getNetworkConfigMock(_ context.Context) (*data.NetworkConfig, error) {
	chainID := "chain ID"
	minGasPrice := uint64(12234)
	minTxVersion := uint32(122)

	return &data.NetworkConfig{
		ChainID:               chainID,
		MinGasPrice:           minGasPrice,
		MinTransactionVersion: minTxVersion,
	}, nil
}
