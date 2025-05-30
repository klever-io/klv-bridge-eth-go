package testsCommon

import "github.com/klever-io/klv-bridge-eth-go/parsers"

// KleverchainCodecStub -
type KleverchainCodecStub struct {
	DecodeProxySCCompleteCallDataCalled  func(buff []byte) (parsers.ProxySCCompleteCallData, error)
	ExtractGasLimitFromRawCallDataCalled func(buff []byte) (uint64, error)
}

// DecodeProxySCCompleteCallData -
func (stub *KleverchainCodecStub) DecodeProxySCCompleteCallData(buff []byte) (parsers.ProxySCCompleteCallData, error) {
	if stub.DecodeProxySCCompleteCallDataCalled != nil {
		return stub.DecodeProxySCCompleteCallDataCalled(buff)
	}

	return parsers.ProxySCCompleteCallData{}, nil
}

// ExtractGasLimitFromRawCallData -
func (stub *KleverchainCodecStub) ExtractGasLimitFromRawCallData(buff []byte) (uint64, error) {
	if stub.ExtractGasLimitFromRawCallDataCalled != nil {
		return stub.ExtractGasLimitFromRawCallDataCalled(buff)
	}

	return 0, nil
}

// IsInterfaceNil -
func (stub *KleverchainCodecStub) IsInterfaceNil() bool {
	return stub == nil
}
