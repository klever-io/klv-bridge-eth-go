package wallet

type Wallet interface {
	PrivateKey() []byte
	PublicKey() []byte
	Sign(msg []byte) ([]byte, error)
	SignHex(msg string) ([]byte, error)
}
