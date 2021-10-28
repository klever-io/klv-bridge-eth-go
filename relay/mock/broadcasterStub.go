package mock

// BroadcasterStub -
type BroadcasterStub struct {
	BroadcastSignatureCalled func(signature []byte)
	BroadcastJoinTopicCalled func()
	ClearSignaturesCalled    func()
	SignaturesCalled         func() [][]byte
	SortedPublicKeysCalled   func() [][]byte
	RegisterOnTopicsCalled   func() error
	CloseCalled              func() error
}

// BroadcastSignature -
func (bs *BroadcasterStub) BroadcastSignature(signature []byte) {
	if bs.BroadcastSignatureCalled != nil {
		bs.BroadcastSignatureCalled(signature)
	}
}

// BroadcastJoinTopic -
func (bs *BroadcasterStub) BroadcastJoinTopic() {
	if bs.BroadcastJoinTopicCalled != nil {
		bs.BroadcastJoinTopicCalled()
	}
}

// ClearSignatures -
func (bs *BroadcasterStub) ClearSignatures() {
	if bs.ClearSignaturesCalled != nil {
		bs.ClearSignaturesCalled()
	}
}

// Signatures -
func (bs *BroadcasterStub) Signatures() [][]byte {
	if bs.SignaturesCalled != nil {
		return bs.SignaturesCalled()
	}

	return make([][]byte, 0)
}

// SortedPublicKeys -
func (bs *BroadcasterStub) SortedPublicKeys() [][]byte {
	if bs.SortedPublicKeysCalled != nil {
		return bs.SortedPublicKeysCalled()
	}

	return make([][]byte, 0)
}

// RegisterOnTopics -
func (bs *BroadcasterStub) RegisterOnTopics() error {
	if bs.RegisterOnTopicsCalled != nil {
		return bs.RegisterOnTopicsCalled()
	}

	return nil
}

// Close -
func (bs *BroadcasterStub) Close() error {
	if bs.CloseCalled() != nil {
		return bs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (bs *BroadcasterStub) IsInterfaceNil() bool {
	return bs == nil
}
