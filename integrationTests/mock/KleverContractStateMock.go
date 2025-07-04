package mock

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	factoryHasher "github.com/klever-io/klever-go/crypto/hashing/factory"
	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/data/vm"
	"github.com/klever-io/klever-go/tools"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
)

type kleverBlockchainProposedStatus struct {
	BatchId  *big.Int
	Statuses []byte
}

type kleverBlockchainProposedTransfer struct {
	BatchId   *big.Int
	Transfers []Transfer
}

// Transfer -
type Transfer struct {
	From   []byte
	To     []byte
	Token  string
	Amount *big.Int
	Nonce  *big.Int
	Data   []byte
}

// KleverBlockchainPendingBatch -
type KleverBlockchainPendingBatch struct {
	Nonce                    *big.Int
	KleverBlockchainDeposits []KleverBlockchainDeposit
}

// KleverBlockchainDeposit -
type KleverBlockchainDeposit struct {
	From         address.Address
	To           common.Address
	Ticker       string
	Amount       *big.Int
	DepositNonce uint64
}

// kleverBlockchainContractStateMock is not concurrent safe
type kleverBlockchainContractStateMock struct {
	*tokensRegistryMock
	proposedStatus                   map[string]*kleverBlockchainProposedStatus   // store them uniquely by their hash
	proposedTransfers                map[string]*kleverBlockchainProposedTransfer // store them uniquely by their hash
	signedActionIDs                  map[string]map[string]struct{}
	GetStatusesAfterExecutionHandler func() []byte
	ProcessFinishedHandler           func()
	relayers                         [][]byte
	performedAction                  *big.Int
	pendingBatch                     *KleverBlockchainPendingBatch
	quorum                           int
	lastExecutedEthBatchId           uint64
	lastExecutedEthTxId              uint64

	ProposeMultiTransferKdaBatchCalled func()
}

func newKleverBlockchainContractStateMock() *kleverBlockchainContractStateMock {
	mock := &kleverBlockchainContractStateMock{
		tokensRegistryMock: &tokensRegistryMock{},
	}
	mock.cleanState()
	mock.clearTokens()

	return mock
}

// Clean -
func (mock *kleverBlockchainContractStateMock) cleanState() {
	mock.proposedStatus = make(map[string]*kleverBlockchainProposedStatus)
	mock.proposedTransfers = make(map[string]*kleverBlockchainProposedTransfer)
	mock.signedActionIDs = make(map[string]map[string]struct{})
	mock.performedAction = nil
	mock.pendingBatch = nil
}

func (mock *kleverBlockchainContractStateMock) processTransaction(tx *transaction.Transaction) {
	dataSplit := strings.Split(string(tx.GetRawData().GetData()[0]), "@")
	funcName := dataSplit[0]
	switch funcName {
	case "proposeKdaSafeSetCurrentTransactionBatchStatus":
		mock.proposeKdaSafeSetCurrentTransactionBatchStatus(dataSplit, tx)

		return
	case "proposeMultiTransferKdaBatch":
		mock.proposeMultiTransferKdaBatch(dataSplit, tx)
		return
	case "sign":
		mock.sign(dataSplit, tx)
		return
	case "performAction":
		mock.performAction(dataSplit, tx)

		if mock.ProcessFinishedHandler != nil {
			mock.ProcessFinishedHandler()
		}
		return
	}

	panic("can not execute transaction that calls function: " + funcName)
}

func (mock *kleverBlockchainContractStateMock) proposeKdaSafeSetCurrentTransactionBatchStatus(dataSplit []string, _ *transaction.Transaction) {
	status, hash := mock.createProposedStatus(dataSplit)

	mock.proposedStatus[hash] = status
}

func (mock *kleverBlockchainContractStateMock) proposeMultiTransferKdaBatch(dataSplit []string, _ *transaction.Transaction) {
	transfer, hash := mock.createProposedTransfer(dataSplit)

	mock.proposedTransfers[hash] = transfer

	if mock.ProposeMultiTransferKdaBatchCalled != nil {
		mock.ProposeMultiTransferKdaBatchCalled()
	}
}

func (mock *kleverBlockchainContractStateMock) createProposedStatus(dataSplit []string) (*kleverBlockchainProposedStatus, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	status := &kleverBlockchainProposedStatus{
		BatchId:  big.NewInt(0).SetBytes(buff),
		Statuses: make([]byte, 0),
	}

	for i := 2; i < len(dataSplit); i++ {
		stat, errDecode := hex.DecodeString(dataSplit[i])
		if errDecode != nil {
			panic(errDecode)
		}

		status.Statuses = append(status.Statuses, stat[0])
	}

	if len(status.Statuses) != len(mock.pendingBatch.KleverBlockchainDeposits) {
		panic(fmt.Sprintf("different number of statuses fetched while creating proposed status: provided %d, existing %d",
			len(status.Statuses), len(mock.pendingBatch.KleverBlockchainDeposits)))
	}

	hasher, err := factoryHasher.NewHasher("blake2b")
	if err != nil {
		panic(err)
	}

	hash, err := tools.CalculateHash(integrationTests.TestMarshalizer, hasher, status)
	if err != nil {
		panic(err)
	}

	return status, string(hash)
}

func (mock *kleverBlockchainContractStateMock) createProposedTransfer(dataSplit []string) (*kleverBlockchainProposedTransfer, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	transfer := &kleverBlockchainProposedTransfer{
		BatchId: big.NewInt(0).SetBytes(buff),
	}

	currentIndex := 2
	for currentIndex < len(dataSplit) {
		from, errDecode := hex.DecodeString(dataSplit[currentIndex])
		if errDecode != nil {
			panic(errDecode)
		}

		to, errDecode := hex.DecodeString(dataSplit[currentIndex+1])
		if errDecode != nil {
			panic(errDecode)
		}

		amountBytes, errDecode := hex.DecodeString(dataSplit[currentIndex+3])
		if errDecode != nil {
			panic(errDecode)
		}

		nonceBytes, errDecode := hex.DecodeString(dataSplit[currentIndex+4])
		if errDecode != nil {
			panic(errDecode)
		}

		dataBytes, errDecode := hex.DecodeString(dataSplit[currentIndex+5])
		if errDecode != nil {
			panic(errDecode)
		}

		t := Transfer{
			From:   from,
			To:     to,
			Token:  dataSplit[currentIndex+2],
			Amount: big.NewInt(0).SetBytes(amountBytes),
			Nonce:  big.NewInt(0).SetBytes(nonceBytes),
			Data:   dataBytes,
		}

		indexIncrementValue := 6
		transfer.Transfers = append(transfer.Transfers, t)
		currentIndex += indexIncrementValue
	}

	hasher, err := factoryHasher.NewHasher("blake2b")
	if err != nil {
		panic(err)
	}

	hash, err := tools.CalculateHash(integrationTests.TestMarshalizer, hasher, transfer)
	if err != nil {
		panic(err)
	}

	actionID := HashToActionID(string(hash))
	integrationTests.Log.Debug("actionID for createProposedTransfer", "value", actionID.String())

	return transfer, string(hash)
}

func (mock *kleverBlockchainContractStateMock) processVmRequests(vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
	if vmRequest == nil {
		panic("vmRequest is nil")
	}

	switch vmRequest.FuncName {
	case "wasTransferActionProposed":
		return mock.vmRequestwasTransferActionProposed(vmRequest), nil
	case "getActionIdForTransferBatch":
		return mock.vmRequestGetActionIdForTransferBatch(vmRequest), nil
	case "wasSetCurrentTransactionBatchStatusActionProposed":
		return mock.vmRequestWasSetCurrentTransactionBatchStatusActionProposed(vmRequest), nil
	case "getStatusesAfterExecution":
		return mock.vmRequestGetStatusesAfterExecution(vmRequest), nil
	case "getActionIdForSetCurrentTransactionBatchStatus":
		return mock.vmRequestGetActionIdForSetCurrentTransactionBatchStatus(vmRequest), nil
	case "wasActionExecuted":
		return mock.vmRequestWasActionExecuted(vmRequest), nil
	case "quorumReached":
		return mock.vmRequestQuorumReached(vmRequest), nil
	case "getTokenIdForErc20Address":
		return mock.vmRequestGetTokenIdForErc20Address(vmRequest), nil
	case "getErc20AddressForTokenId":
		return mock.vmRequestGetErc20AddressForTokenId(vmRequest), nil
	case "getCurrentTxBatch":
		return mock.vmRequestGetCurrentPendingBatch(vmRequest), nil
	case "getBatch":
		return mock.vmRequestGetBatch(vmRequest), nil
	case "getAllStakedRelayers":
		return mock.vmRequestGetAllStakedRelayers(vmRequest), nil
	case "getLastExecutedEthBatchId":
		return mock.vmRequestGetLastExecutedEthBatchId(vmRequest), nil
	case "getLastExecutedEthTxId":
		return mock.vmRequestGetLastExecutedEthTxId(vmRequest), nil
	case "signed":
		return mock.vmRequestSigned(vmRequest), nil
	case "isPaused":
		return mock.vmRequestIsPaused(vmRequest), nil
	case "isMintBurnToken":
		return mock.vmRequestIsMintBurnToken(vmRequest), nil
	case "isNativeToken":
		return mock.vmRequestIsNativeToken(vmRequest), nil
	case "getTotalBalances":
		return mock.vmRequestGetTotalBalances(vmRequest), nil
	case "getMintBalances":
		return mock.vmRequestGetMintBalances(vmRequest), nil
	case "getBurnBalances":
		return mock.vmRequestGetBurnBalances(vmRequest), nil
	case "getLastBatchId":
		return mock.vmRequestGetLastBatchId(vmRequest), nil
	}

	panic("unimplemented function: " + vmRequest.FuncName)
}

func (mock *kleverBlockchainContractStateMock) vmRequestWasSetCurrentTransactionBatchStatusActionProposed(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetActionIdForSetCurrentTransactionBatchStatus(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]
	if !found {
		return createNokVmResponse(fmt.Errorf("proposed status not found for hash %s", hex.EncodeToString([]byte(hash))))
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestwasTransferActionProposed(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetActionIdForTransferBatch(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]
	if !found {
		// return action ID == 0 in case there is no such transfer proposed
		return createOkVmResponse([][]byte{big.NewInt(0).Bytes()})
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetStatusesAfterExecution(_ *models.VmValueRequest) *models.VmValuesResponseData {
	statuses := mock.GetStatusesAfterExecutionHandler()

	args := [][]byte{BoolToByteSlice(true)} // batch finished
	for _, stat := range statuses {
		args = append(args, []byte{stat})
	}

	return createOkVmResponse(args)
}

func (mock *kleverBlockchainContractStateMock) sign(dataSplit []string, tx *transaction.Transaction) {
	actionID := getBigIntFromString(dataSplit[1])
	if !mock.actionIDExists(actionID) {
		panic(fmt.Sprintf("attempted to sign on a missing action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		m = make(map[string]struct{})
		mock.signedActionIDs[actionID.String()] = m
	}
	m[string(tx.GetRawData().GetSender())] = struct{}{}
}

func (mock *kleverBlockchainContractStateMock) performAction(dataSplit []string, _ *transaction.Transaction) {
	actionID := getBigIntFromString(dataSplit[1])
	if !mock.actionIDExists(actionID) {
		panic(fmt.Sprintf("attempted to perform on a missing action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		panic(fmt.Sprintf("attempted to perform on a not signed action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	if len(m) >= mock.quorum {
		mock.performedAction = actionID
	}
}

func (mock *kleverBlockchainContractStateMock) vmRequestWasActionExecuted(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	actionID := getBigIntFromString(vmRequest.Args[0])

	if mock.performedAction == nil {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	actionProposed := actionID.Cmp(mock.performedAction) == 0

	return createOkVmResponse([][]byte{BoolToByteSlice(actionProposed)})
}

func (mock *kleverBlockchainContractStateMock) actionIDExists(actionID *big.Int) bool {
	for hash := range mock.proposedTransfers {
		existingActionID := HashToActionID(hash)
		if existingActionID.Cmp(actionID) == 0 {
			return true
		}
	}

	for hash := range mock.proposedStatus {
		existingActionID := HashToActionID(hash)
		if existingActionID.Cmp(actionID) == 0 {
			return true
		}
	}

	return false
}

func (mock *kleverBlockchainContractStateMock) vmRequestQuorumReached(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	actionID := getBigIntFromString(vmRequest.Args[0])
	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	quorumReached := len(m) >= mock.quorum

	return createOkVmResponse([][]byte{BoolToByteSlice(quorumReached)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetTokenIdForErc20Address(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := common.HexToAddress(vmRequest.Args[0])

	return createOkVmResponse([][]byte{[]byte(mock.getTicker(address))})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetErc20AddressForTokenId(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getErc20Address(address).Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetAllStakedRelayers(_ *models.VmValueRequest) *models.VmValuesResponseData {
	return createOkVmResponse(mock.relayers)
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetLastExecutedEthBatchId(_ *models.VmValueRequest) *models.VmValuesResponseData {
	val := big.NewInt(int64(mock.lastExecutedEthBatchId))

	return createOkVmResponse([][]byte{val.Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetLastExecutedEthTxId(_ *models.VmValueRequest) *models.VmValuesResponseData {
	val := big.NewInt(int64(mock.lastExecutedEthTxId))

	return createOkVmResponse([][]byte{val.Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetCurrentPendingBatch(_ *models.VmValueRequest) *models.VmValuesResponseData {
	if mock.pendingBatch == nil {
		return createOkVmResponse(make([][]byte, 0))
	}

	return mock.responseWithPendingBatch()
}

func (mock *kleverBlockchainContractStateMock) responseWithPendingBatch() *models.VmValuesResponseData {
	args := [][]byte{mock.pendingBatch.Nonce.Bytes()} // first non-empty slice
	for _, deposit := range mock.pendingBatch.KleverBlockchainDeposits {
		args = append(args, make([]byte, 0)) // mocked block nonce
		args = append(args, big.NewInt(0).SetUint64(deposit.DepositNonce).Bytes())
		args = append(args, deposit.From.Bytes())
		args = append(args, deposit.To.Bytes())
		args = append(args, []byte(deposit.Ticker))
		args = append(args, deposit.Amount.Bytes())
	}
	return createOkVmResponse(args)
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetBatch(request *models.VmValueRequest) *models.VmValuesResponseData {
	if mock.pendingBatch == nil {
		return createOkVmResponse(make([][]byte, 0))
	}

	nonce := getBigIntFromString(request.Args[0])
	if nonce.Cmp(mock.pendingBatch.Nonce) == 0 {
		return mock.responseWithPendingBatch()
	}

	return createOkVmResponse(make([][]byte, 0))
}

func (mock *kleverBlockchainContractStateMock) setPendingBatch(pendingBatch *KleverBlockchainPendingBatch) {
	mock.pendingBatch = pendingBatch
}

func (mock *kleverBlockchainContractStateMock) vmRequestSigned(request *models.VmValueRequest) *models.VmValuesResponseData {
	hexAddress := request.Args[0]
	actionID := getBigIntFromString(request.Args[1])

	actionIDMap, found := mock.signedActionIDs[actionID.String()]
	if !found {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	addressBytes, err := hex.DecodeString(hexAddress)
	if err != nil {
		panic(err)
	}

	address, err := address.NewAddressFromBytes(addressBytes)
	if err != nil {
		panic(err)
	}
	bech32Address := address.Bech32()
	_, found = actionIDMap[bech32Address]
	if !found {
		log.Error("action ID not found", "address", bech32Address)
	}

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestIsPaused(_ *models.VmValueRequest) *models.VmValuesResponseData {
	return createOkVmResponse([][]byte{BoolToByteSlice(false)})
}

func (mock *kleverBlockchainContractStateMock) vmRequestIsMintBurnToken(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{BoolToByteSlice(mock.isMintBurnToken(address))})
}

func (mock *kleverBlockchainContractStateMock) vmRequestIsNativeToken(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{BoolToByteSlice(mock.isNativeToken(address))})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetTotalBalances(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getTotalBalances(address).Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetMintBalances(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getMintBalances(address).Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetBurnBalances(vmRequest *models.VmValueRequest) *models.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getBurnBalances(address).Bytes()})
}

func (mock *kleverBlockchainContractStateMock) vmRequestGetLastBatchId(_ *models.VmValueRequest) *models.VmValuesResponseData {
	if mock.pendingBatch == nil {
		return createOkVmResponse([][]byte{big.NewInt(0).Bytes()})
	}
	return createOkVmResponse([][]byte{mock.pendingBatch.Nonce.Bytes()})
}

func getBigIntFromString(data string) *big.Int {
	buff, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}

	return big.NewInt(0).SetBytes(buff)
}

func createOkVmResponse(args [][]byte) *models.VmValuesResponseData {
	return &models.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnData: args,
			ReturnCode: "Ok",
		},
	}
}

func createNokVmResponse(err error) *models.VmValuesResponseData {
	return &models.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnCode:    "nok",
			ReturnMessage: err.Error(),
		},
	}
}
