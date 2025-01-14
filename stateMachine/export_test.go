package stateMachine

import "github.com/klever-io/klv-bridge-eth-go/core"

// GetCurrentStep -
func (sm *stateMachine) GetCurrentStepIdentifier() core.StepIdentifier {
	return sm.currentStep.Identifier()
}
