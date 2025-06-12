//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromKcFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToKlvDone bool
	kdaToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromKcFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.kdaToEthDone && flow.ethToKlvDone {
		return true
	}

	isTransferDoneFromKc := flow.setup.IsTransferDoneFromKc(flow.tokens...)
	if !flow.kdaToEthDone && isTransferDoneFromKc {
		flow.kdaToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Klever Blockchain->Ethereum transfer finished, now sending back to Klever Blockchain..."))

		flow.setup.EthereumHandler.SendFromEthereumToKc(flow.setup.Ctx, flow.setup.KcHandler.TestCallerAddress, flow.tokens...)
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

func (flow *startsFromKcFlow) areTokensFullyRefunded() bool {
	if len(flow.tokens) == 0 {
		return true
	}
	if !flow.ethToKlvDone {
		return false // regular flow is not completed
	}

	return flow.setup.IsTransferDoneFromEthereumWithRefund(flow.tokens...)
}
