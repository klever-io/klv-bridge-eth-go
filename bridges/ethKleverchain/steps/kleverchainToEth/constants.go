package kleverchaintoeth

const (
	// GettingPendingBatchFromKleverchain is the step identifier for fetching the pending batch from the Kleverchain
	GettingPendingBatchFromKleverchain = "get pending batch from Kleverchain"

	// SigningProposedTransferOnEthereum is the step identifier for signing proposed transfer
	SigningProposedTransferOnEthereum = "sign proposed transfer"

	// WaitingForQuorumOnTransfer is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnTransfer = "wait for quorum on transfer"

	// PerformingTransfer is the step identifier for performing the transfer on Ethereum
	PerformingTransfer = "perform transfer"

	// WaitingTransferConfirmation is the step identifier for waiting the transfer confirmation on Ethereum
	WaitingTransferConfirmation = "wait transfer confirmating"

	// ResolvingSetStatusOnKleverchain is the step identifier for resolving set status on Kleverchain
	ResolvingSetStatusOnKleverchain = "resolve set status"

	// ProposingSetStatusOnKleverchain is the step idetifier for proposing set status action on Kleverchain
	ProposingSetStatusOnKleverchain = "propose set status"

	// SigningProposedSetStatusOnKleverchain is the step identifier for signing proposed set status action
	SigningProposedSetStatusOnKleverchain = "sign proposed set status"

	// WaitingForQuorumOnSetStatus is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnSetStatus = "wait for quorum on set status"

	// PerformingSetStatus is the step identifier for performing the set status action on Kleverchain
	PerformingSetStatus = "perform set status"

	// NumSteps indicates how many steps the state machine for Kleverchain -> Ethereum flow has
	NumSteps = 10
)
