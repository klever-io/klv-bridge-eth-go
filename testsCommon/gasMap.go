package testsCommon

import (
	"github.com/klever-io/klv-bridge-eth-go/config"
)

// CreateTestKleverGasMap will create a testing gas map for Klever client
func CreateTestKleverGasMap() config.KleverGasMapConfig {
	return config.KleverGasMapConfig{
		Sign:                   101,
		ProposeTransferBase:    102,
		ProposeTransferForEach: 103,
		ProposeStatusBase:      104,
		ProposeStatusForEach:   105,
		PerformActionBase:      106,
		PerformActionForEach:   107,
		ScCallPerByte:          108,
		ScCallPerformForEach:   109,
	}
}
