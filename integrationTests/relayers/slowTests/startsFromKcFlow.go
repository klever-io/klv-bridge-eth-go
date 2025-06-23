//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromKCFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToKlvDone bool
	kdaToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromKCFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.kdaToEthDone && flow.ethToKlvDone {
		return true
	}

	isTransferDoneFromKC := flow.setup.IsTransferDoneFromKC(flow.tokens...)
	if !flow.kdaToEthDone && isTransferDoneFromKC {
		flow.kdaToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Klever Blockchain->Ethereum transfer finished, now sending back to Klever Blockchain..."))

		flow.setup.EthereumHandler.SendFromEthereumToKC(flow.setup.Ctx, flow.setup.KCHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.kdaToEthDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromEthereum := flow.setup.IsTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToKlvDone && isTransferDoneFromEthereum {
		flow.ethToKlvDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Klever Blockchain<->Ethereum from Klever Blockchain transfers done"))
		return true
	}

	return false
}

func (flow *startsFromKCFlow) areTokensFullyRefunded() bool {
	if len(flow.tokens) == 0 {
		return true
	}
	if !flow.ethToKlvDone {
		return false // regular flow is not completed
	}

	return flow.setup.IsTransferDoneFromEthereumWithRefund(flow.tokens...)
}
