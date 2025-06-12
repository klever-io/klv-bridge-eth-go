package kleverchaintoeth

import (
	"testing"

	ethKleverchain "github.com/klever-io/klv-bridge-eth-go/bridges/ethKleverchain"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSteps_Errors(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(nil)

	assert.Nil(t, steps)
	assert.Equal(t, ethKleverchain.ErrNilExecutor, err)
}

func TestCreateSteps_ShouldWork(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(bridgeTests.NewBridgeExecutorStub())

	require.NotNil(t, steps)
	require.Nil(t, err)
	require.Equal(t, NumSteps, len(steps))
}
