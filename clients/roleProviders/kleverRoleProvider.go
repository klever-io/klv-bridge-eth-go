package roleproviders

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgsKleverRoleProvider is the argument for the Klever Blockchain role provider constructor
type ArgsKleverRoleProvider struct {
	DataGetter DataGetter
	Log        logger.Logger
}

type kleverRoleProvider struct {
	dataGetter           DataGetter
	log                  logger.Logger
	whitelistedAddresses map[string]struct{}
	mut                  sync.RWMutex
}

// NewKleverRoleProvider creates a new kleverRoleProvider instance able to fetch the whitelisted addresses
func NewKleverRoleProvider(args ArgsKleverRoleProvider) (*kleverRoleProvider, error) {
	err := checkKleverRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	krp := &kleverRoleProvider{
		dataGetter:           args.DataGetter,
		log:                  args.Log,
		whitelistedAddresses: make(map[string]struct{}),
	}

	return krp, nil
}

func checkKleverRoleProviderSpecificArgs(args ArgsKleverRoleProvider) error {
	if check.IfNil(args.DataGetter) {
		return clients.ErrNilDataGetter
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}

	return nil
}

// Execute will fetch the available relayers and store them in the inner map
func (krp *kleverRoleProvider) Execute(ctx context.Context) error {
	results, err := krp.dataGetter.GetAllStakedRelayers(ctx)
	if err != nil {
		return err
	}

	return krp.processResults(results)
}

func (krp *kleverRoleProvider) processResults(results [][]byte) error {
	currentList := make([]string, 0, len(results))
	temporaryMap := make(map[string]struct{})

	for i, result := range results {
		address, err := address.NewAddressFromBytes(result)
		if err != nil {
			return fmt.Errorf("%w for index %d, malformed address: %s", ErrInvalidAddressBytes, i, hex.EncodeToString(result))
		}

		isValid := address.IsValid()
		if !isValid {
			return fmt.Errorf("%w for index %d, malformed address: %s", ErrInvalidAddressBytes, i, hex.EncodeToString(result))
		}

		bech32Address := address.Bech32()
		currentList = append(currentList, bech32Address)
		temporaryMap[string(address.Bytes())] = struct{}{}
	}

	krp.mut.Lock()
	krp.whitelistedAddresses = temporaryMap
	krp.mut.Unlock()

	krp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))

	return nil
}

// IsWhitelisted returns true if the non-nil address provided is whitelisted or not
func (krp *kleverRoleProvider) IsWhitelisted(address address.Address) bool {
	if check.IfNil(address) {
		return false
	}

	krp.mut.RLock()
	defer krp.mut.RUnlock()

	_, exists := krp.whitelistedAddresses[string(address.Bytes())]

	return exists
}

// SortedPublicKeys will return all the sorted public keys
func (krp *kleverRoleProvider) SortedPublicKeys() [][]byte {
	krp.mut.RLock()
	defer krp.mut.RUnlock()

	sortedPublicKeys := make([][]byte, 0, len(krp.whitelistedAddresses))
	for addr := range krp.whitelistedAddresses {
		sortedPublicKeys = append(sortedPublicKeys, []byte(addr))
	}

	sort.Slice(sortedPublicKeys, func(i, j int) bool {
		return bytes.Compare(sortedPublicKeys[i], sortedPublicKeys[j]) < 0
	})
	return sortedPublicKeys
}

// IsInterfaceNil returns true if there is no value under the interface
func (krp *kleverRoleProvider) IsInterfaceNil() bool {
	return krp == nil
}
