//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromEthereumEdgecaseFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToKlvDone bool
	kdaToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromEthereumEdgecaseFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.kdaToEthDone && flow.ethToKlvDone {
		return true
	}

	isTransferDoneFromEthereum := flow.setup.IsTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToKlvDone && isTransferDoneFromEthereum {
		flow.ethToKlvDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Ethereum->Kleverchain transfer finished, now sending back to Ethereum & another round from Ethereum..."))

		flow.setup.SendFromKleverchainToEthereum(flow.tokens...)
		flow.setup.EthereumHandler.SendFromEthereumToKleverchain(flow.setup.Ctx, flow.setup.KleverchainHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.ethToKlvDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromKleverchain := flow.setup.IsTransferDoneFromKleverchain(flow.tokens...)
	if !flow.kdaToEthDone && isTransferDoneFromKleverchain {
		flow.kdaToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Kleverchain<->Ethereum from Ethereum transfers done"))
		return true
	}

	return false
}
