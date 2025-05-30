//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromKleverchainFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToKlvDone bool
	kdaToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromKleverchainFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.kdaToEthDone && flow.ethToKlvDone {
		return true
	}

	isTransferDoneFromKleverchain := flow.setup.IsTransferDoneFromKleverchain(flow.tokens...)
	if !flow.kdaToEthDone && isTransferDoneFromKleverchain {
		flow.kdaToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Kleverchain->Ethereum transfer finished, now sending back to Kleverchain..."))

		flow.setup.EthereumHandler.SendFromEthereumToKleverchain(flow.setup.Ctx, flow.setup.KleverchainHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.kdaToEthDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromEthereum := flow.setup.IsTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToKlvDone && isTransferDoneFromEthereum {
		flow.ethToKlvDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Kleverchain<->Ethereum from Kleverchain transfers done"))
		return true
	}

	return false
}

func (flow *startsFromKleverchainFlow) areTokensFullyRefunded() bool {
	if len(flow.tokens) == 0 {
		return true
	}
	if !flow.ethToKlvDone {
		return false // regular flow is not completed
	}

	return flow.setup.IsTransferDoneFromEthereumWithRefund(flow.tokens...)
}
