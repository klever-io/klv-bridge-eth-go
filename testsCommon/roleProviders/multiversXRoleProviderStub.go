package roleproviders

import (
	"github.com/klever-io/klever-go-sdk/core/address"
)

// MultiversXRoleProviderStub -
type KleverRoleProviderStub struct {
	IsWhitelistedCalled func(address address.Address) bool
}

// IsWhitelisted -
func (stub *KleverRoleProviderStub) IsWhitelisted(address address.Address) bool {
	if stub.IsWhitelistedCalled != nil {
		return stub.IsWhitelistedCalled(address)
	}

	return true
}

// IsInterfaceNil -
func (stub *KleverRoleProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
