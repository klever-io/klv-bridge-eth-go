package ethtokleverchain

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from Ethereum"

	// ProposingTransferOnKleverchain is the step identifier for proposing transfer on Kleverchain
	ProposingTransferOnKleverchain = "propose transfer"

	// SigningProposedTransferOnKleverchain is the step identifier for signing proposed transfer
	SigningProposedTransferOnKleverchain = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on Kleverchain
	PerformingActionID = "perform action"

	// NumSteps indicates how many steps the state machine for Ethereum -> Kleverchain flow has
	NumSteps = 5
)
