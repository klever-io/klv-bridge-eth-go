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
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Ethereum->Klever Blockchain transfer finished, now sending back to Ethereum & another round from Ethereum..."))

		flow.setup.SendFromKCToEthereum(flow.tokens...)
		flow.setup.EthereumHandler.SendFromEthereumToKC(flow.setup.Ctx, flow.setup.KCHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.ethToKlvDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromKC := flow.setup.IsTransferDoneFromKC(flow.tokens...)
	if !flow.kdaToEthDone && isTransferDoneFromKC {
		flow.kdaToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Klever Blockchain<->Ethereum from Ethereum transfers done"))
		return true
	}

	return false
}
