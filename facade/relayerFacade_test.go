package facade

import (
	"errors"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/status"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArguments() ArgsRelayerFacade {
	return ArgsRelayerFacade{
		MetricsHolder: status.NewMetricsHolder(),
		ApiInterface:  core.WebServerOffString,
		PprofEnabled:  true,
	}
}

func TestNewRelayerFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil metrics holder should error", func(t *testing.T) {
		args := createMockArguments()
		args.MetricsHolder = nil

		facade, err := NewRelayerFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrNilMetricsHolder))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArguments()

		facade, err := NewRelayerFacade(args)
		assert.False(t, check.IfNil(facade))
		assert.Nil(t, err)
	})
}

func TestRelayerFacade_Getters(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	facade, _ := NewRelayerFacade(args)

	assert.Equal(t, args.ApiInterface, facade.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facade.PprofEnabled())
}

func TestRelayerFacade_GetMetrics(t *testing.T) {
	t.Parallel()

	sh1 := testsCommon.NewStatusHandlerMock("mock1")
	sh2 := testsCommon.NewStatusHandlerMock("mock2")
	sh2.SetStringMetric("metric1", "value1")
	metricHolder := status.NewMetricsHolder()
	errSetup := metricHolder.AddStatusHandler(sh1)
	require.Nil(t, errSetup)
	errSetup = metricHolder.AddStatusHandler(sh2)
	require.Nil(t, errSetup)

	t.Run("name not found should error", func(t *testing.T) {
		args := createMockArguments()
		args.MetricsHolder = metricHolder
		facade, _ := NewRelayerFacade(args)

		response, err := facade.GetMetrics("not-found")
		require.Nil(t, response)
		require.NotNil(t, err)
	})
	t.Run("name exists should return the available metrics", func(t *testing.T) {
		args := createMockArguments()
		args.MetricsHolder = metricHolder
		facade, _ := NewRelayerFacade(args)

		response, err := facade.GetMetrics("mock2")
		require.Nil(t, err)
		require.Equal(t, sh2.GetAllMetrics(), response)
	})
}

func TestRelayerFacade_GetMetricsList(t *testing.T) {
	t.Parallel()

	sh1 := testsCommon.NewStatusHandlerMock("mock1")
	sh2 := testsCommon.NewStatusHandlerMock("mock2")
	sh2.SetStringMetric("metric1", "value1")
	metricHolder := status.NewMetricsHolder()
	errSetup := metricHolder.AddStatusHandler(sh1)
	require.Nil(t, errSetup)
	errSetup = metricHolder.AddStatusHandler(sh2)
	require.Nil(t, errSetup)

	args := createMockArguments()
	args.MetricsHolder = metricHolder
	facade, _ := NewRelayerFacade(args)

	response := facade.GetMetricsList()
	expected := make(core.GeneralMetrics)
	expected[availableMetrics] = []string{"mock1", "mock2"}
	assert.Equal(t, expected, response)
}
