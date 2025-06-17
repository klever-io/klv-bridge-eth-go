package ethtokc

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from Ethereum"

	// ProposingTransferOnKc is the step identifier for proposing transfer on Kc
	ProposingTransferOnKc = "propose transfer"

	// SigningProposedTransferOnKc is the step identifier for signing proposed transfer
	SigningProposedTransferOnKc = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on Kc
	PerformingActionID = "perform action"

	// NumSteps indicates how many steps the state machine for Ethereum -> Kc flow has
	NumSteps = 5
)
