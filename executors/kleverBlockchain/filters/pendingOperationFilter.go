package filters

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	wildcardString   = "*"
	emptyString      = ""
	ethAddressPrefix = "0x"
)

var ethWildcardString = ""

func init() {
	var ethAddressWildcard = common.Address{}
	ethAddressWildcard.SetBytes([]byte(wildcardString))
	ethWildcardString = ethAddressWildcard.String()
}

type pendingOperationFilter struct {
	allowedEthAddresses []string
	deniedEthAddresses  []string
	allowedKlvAddresses []string
	deniedKlvAddresses  []string
	allowedTokens       []string
	deniedTokens        []string
}

// NewPendingOperationFilter creates a new instance of type pendingOperationFilter
func NewPendingOperationFilter(cfg config.PendingOperationsFilterConfig, log logger.Logger) (*pendingOperationFilter, error) {
	if check.IfNil(log) {
		return nil, errNilLogger
	}
	if len(cfg.AllowedKlvAddresses)+len(cfg.AllowedEthAddresses)+len(cfg.AllowedTokens) == 0 {
		return nil, errNoItemsAllowed
	}

	filter := &pendingOperationFilter{}
	err := filter.parseConfigs(cfg)
	if err != nil {
		return nil, err
	}

	err = filter.checkLists()
	if err != nil {
		return nil, err
	}

	log.Info("NewPendingOperationFilter config options",
		"DeniedEthAddresses", strings.Join(filter.deniedEthAddresses, ", "),
		"DeniedKlvAddresses", strings.Join(filter.deniedKlvAddresses, ", "),
		"DeniedTokens", strings.Join(filter.deniedTokens, ", "),
		"AllowedEthAddresses", strings.Join(filter.allowedEthAddresses, ", "),
		"AllowedKlvAddresses", strings.Join(filter.allowedKlvAddresses, ", "),
		"AllowedTokens", strings.Join(filter.allowedTokens, ", "),
	)

	return filter, nil
}

func (filter *pendingOperationFilter) parseConfigs(cfg config.PendingOperationsFilterConfig) error {
	var err error

	// denied lists do not support wildcard items
	filter.deniedEthAddresses, err = parseList(cfg.DeniedEthAddresses, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedEthAddresses", err)
	}

	filter.deniedKlvAddresses, err = parseList(cfg.DeniedKlvAddresses, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedKlvAddresses", err)
	}

	filter.deniedTokens, err = parseList(cfg.DeniedTokens, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedTokens", err)
	}

	// allowed lists do not support empty items
	filter.allowedEthAddresses, err = parseList(cfg.AllowedEthAddresses, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedEthAddresses", err)
	}

	filter.allowedKlvAddresses, err = parseList(cfg.AllowedKlvAddresses, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedKlvAddresses", err)
	}

	filter.allowedTokens, err = parseList(cfg.AllowedTokens, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedTokens", err)
	}

	return nil
}

func parseList(list []string, unsupportedMarker string) ([]string, error) {
	newList := make([]string, 0, len(list))
	for index, item := range list {
		item = strings.ToLower(item)
		item = strings.Trim(item, "\r\n \t")
		if item == unsupportedMarker {
			return nil, fmt.Errorf("%w %s on item at index %d", errUnsupportedMarker, unsupportedMarker, index)
		}

		newList = append(newList, item)
	}

	return newList, nil
}

func (filter *pendingOperationFilter) checkLists() error {
	err := filter.checkList(filter.allowedEthAddresses, checkEthItemValid)
	if err != nil {
		return fmt.Errorf("%w in list AllowedEthAddresses", err)
	}

	err = filter.checkList(filter.deniedEthAddresses, checkEthItemValid)
	if err != nil {
		return fmt.Errorf("%w in list DeniedEthAddresses", err)
	}

	err = filter.checkList(filter.allowedKlvAddresses, checkKlvItemValid)
	if err != nil {
		return fmt.Errorf("%w in list AllowedKlvAddresses", err)
	}

	err = filter.checkList(filter.deniedKlvAddresses, checkKlvItemValid)
	if err != nil {
		return fmt.Errorf("%w in list DeniedKlvAddresses", err)
	}

	return nil
}

func (filter *pendingOperationFilter) checkList(list []string, checkItem func(item string) error) error {
	for index, item := range list {
		if item == wildcardString {
			continue
		}

		err := checkItem(item)
		if err != nil {
			return fmt.Errorf("%w on item at index %d", err, index)
		}
	}

	return nil
}

func checkKlvItemValid(item string) error {
	_, errNewAddr := address.NewAddress(item)
	return errNewAddr
}

func checkEthItemValid(item string) error {
	if !strings.HasPrefix(item, ethAddressPrefix) {
		return fmt.Errorf("%w (missing %s prefix)", errMissingEthPrefix, ethAddressPrefix)
	}

	return nil
}

// ShouldExecute returns true if the To, From or token are not denied and allowed
func (filter *pendingOperationFilter) ShouldExecute(callData parsers.ProxySCCompleteCallData) bool {
	if check.IfNil(callData.To) {
		return false
	}

	toAddress := callData.To.Bech32()
	isSpecificallyDenied := filter.stringExistsInList(callData.From.String(), filter.deniedEthAddresses, ethWildcardString) ||
		filter.stringExistsInList(toAddress, filter.deniedKlvAddresses, wildcardString) ||
		filter.stringExistsInList(callData.Token, filter.deniedTokens, wildcardString)
	if isSpecificallyDenied {
		return false
	}

	isAllowed := filter.stringExistsInList(callData.From.String(), filter.allowedEthAddresses, ethWildcardString) ||
		filter.stringExistsInList(toAddress, filter.allowedKlvAddresses, wildcardString) ||
		filter.stringExistsInList(callData.Token, filter.allowedTokens, wildcardString)

	return isAllowed
}

func (filter *pendingOperationFilter) stringExistsInList(needle string, haystack []string, wildcardMarker string) bool {
	needle = strings.ToLower(needle)
	wildcardMarker = strings.ToLower(wildcardMarker)

	for _, item := range haystack {
		if item == wildcardMarker {
			return true
		}

		if item == needle {
			return true
		}
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (filter *pendingOperationFilter) IsInterfaceNil() bool {
	return filter == nil
}
