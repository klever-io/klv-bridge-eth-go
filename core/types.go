package core

import (
	"context"
	"time"
)

// StepIdentifier defines a step name
type StepIdentifier string

// MachineStates defines all available steps for a state machine to run
type MachineStates map[StepIdentifier]Step

// Step defines a state machine step
type Step interface {
	Execute(ctx context.Context) (StepIdentifier, error)
	Identifier() StepIdentifier
	IsInterfaceNil() bool
}

// EthGasPriceSelector defines the ethereum gas price selector
type EthGasPriceSelector string

// Timer defines operations related to time
type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
	IsInterfaceNil() bool
}

// BroadcastClient defines a broadcast client that will get notified by the broadcaster
// when new messages arrive. It also should be able to respond with any stored messages it might
// have.
type BroadcastClient interface {
	ProcessNewMessage(msg *SignedMessage, ethMsg *EthereumSignature)
	AllStoredSignatures() []*SignedMessage
	IsInterfaceNil() bool
}

// StatusHandler is able to keep metrics
type StatusHandler interface {
	SetIntMetric(metric string, value int)
	AddIntMetric(metric string, delta int)
	SetStringMetric(metric string, val string)
	GetAllMetrics() GeneralMetrics
	Name() string
	IsInterfaceNil() bool
}

// MetricsHolder represents the component that can hold metrics
type MetricsHolder interface {
	AddStatusHandler(sh StatusHandler) error
	GetAvailableStatusHandlers() []string
	GetAllMetrics(name string) (GeneralMetrics, error)
	IsInterfaceNil() bool
}

// GeneralMetrics represents an objects metrics map
type GeneralMetrics map[string]interface{}

// StringMetrics represents string metrics map
type StringMetrics map[string]string

// IntMetrics represents string metrics map
type IntMetrics map[string]int
