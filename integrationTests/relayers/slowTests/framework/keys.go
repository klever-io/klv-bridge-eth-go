package framework

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	klvCrypto "github.com/klever-io/klever-go/crypto"
	"github.com/stretchr/testify/require"
)

// constants for the keys store
const (
	relayerPemPathFormat         = "klever%d.pem"
	SCCallerFilename             = "scCaller.pem"
	projectedShardForBridgeSetup = byte(0)
	projectedShardForDepositor   = byte(1)
	projectedShardForTestKeys    = byte(2)
)

// KeysHolder holds a 2 pk-sk pairs for both chains
type KeysHolder struct {
	KlvAddress *KlvAddress
	KlvSk      []byte
	EthSK      *ecdsa.PrivateKey
	EthAddress common.Address
}

// KeysStore will hold all the keys used in the test
type KeysStore struct {
	testing.TB
	RelayersKeys   []KeysHolder
	OraclesKeys    []KeysHolder
	SCExecutorKeys KeysHolder
	OwnerKeys      KeysHolder
	DepositorKeys  KeysHolder
	TestKeys       KeysHolder
	workingDir     string
}

const (
	ethOwnerSK     = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	ethDepositorSK = "9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1"
	ethTestSk      = "dafea2c94bfe5d25f1a508808c2bc2c2e6c6f18b6b010fc841d8eb80755ba27a"
)

// NewKeysStore will create a KeysStore instance and generate all keys
func NewKeysStore(
	tb testing.TB,
	workingDir string,
	numRelayers int,
	numOracles int,
) *KeysStore {
	keysStore := &KeysStore{
		TB:             tb,
		RelayersKeys:   make([]KeysHolder, 0, numRelayers),
		SCExecutorKeys: KeysHolder{},
		workingDir:     workingDir,
	}

	keysStore.generateRelayersKeys(numRelayers)
	keysStore.OraclesKeys = keysStore.generateKeys(numOracles, "generated oracle", projectedShardForBridgeSetup)
	keysStore.SCExecutorKeys = keysStore.generateKey("", projectedShardForBridgeSetup)
	keysStore.OwnerKeys = keysStore.generateKey(ethOwnerSK, projectedShardForBridgeSetup)
	log.Info("generated owner",
		"Klv address", keysStore.OwnerKeys.KlvAddress.Bech32(),
		"Eth address", keysStore.OwnerKeys.EthAddress.String())
	keysStore.DepositorKeys = keysStore.generateKey(ethDepositorSK, projectedShardForDepositor)
	keysStore.TestKeys = keysStore.generateKey(ethTestSk, projectedShardForTestKeys)

	filename := path.Join(keysStore.workingDir, SCCallerFilename)
	SaveKlvKey(keysStore, filename, keysStore.SCExecutorKeys)

	return keysStore
}

func (keyStore *KeysStore) generateRelayersKeys(numKeys int) {
	for i := 0; i < numKeys; i++ {
		relayerETHSKBytes, err := os.ReadFile(fmt.Sprintf(relayerETHKeyPathFormat, i))
		require.Nil(keyStore, err)

		relayerKeys := keyStore.generateKey(string(relayerETHSKBytes), projectedShardForBridgeSetup)
		log.Info("generated relayer", "index", i,
			"Klv address", relayerKeys.KlvAddress.Bech32(),
			"Eth address", relayerKeys.EthAddress.String())

		keyStore.RelayersKeys = append(keyStore.RelayersKeys, relayerKeys)

		filename := path.Join(keyStore.workingDir, fmt.Sprintf(relayerPemPathFormat, i))

		SaveKlvKey(keyStore, filename, relayerKeys)
	}
}

func (keyStore *KeysStore) generateKeys(numKeys int, message string, projectedShard byte) []KeysHolder {
	keys := make([]KeysHolder, 0, numKeys)

	for i := 0; i < numKeys; i++ {
		ethPrivateKeyBytes := make([]byte, 32)
		_, _ = rand.Read(ethPrivateKeyBytes)

		key := keyStore.generateKey(hex.EncodeToString(ethPrivateKeyBytes), projectedShard)
		log.Info(message, "index", i,
			"Klv address", key.KlvAddress.Bech32(),
			"Eth address", key.EthAddress.String())

		keys = append(keys, key)
	}

	return keys
}

func (keyStore *KeysStore) generateKey(ethSkHex string, projectedShard byte) KeysHolder {
	var err error

	keys := GenerateKlvPrivatePublicKey(keyStore, projectedShard)
	if len(ethSkHex) == 0 {
		// eth keys not required
		return keys
	}

	keys.EthSK, err = crypto.HexToECDSA(ethSkHex)
	require.Nil(keyStore, err)

	keys.EthAddress = crypto.PubkeyToAddress(keys.EthSK.PublicKey)

	return keys
}

func (keyStore *KeysStore) getAllKeys() []KeysHolder {
	allKeys := make([]KeysHolder, 0, len(keyStore.RelayersKeys)+10)
	allKeys = append(allKeys, keyStore.RelayersKeys...)
	allKeys = append(allKeys, keyStore.OraclesKeys...)
	allKeys = append(allKeys, keyStore.SCExecutorKeys, keyStore.OwnerKeys, keyStore.DepositorKeys, keyStore.TestKeys)

	return allKeys
}

// WalletsToFundOnEthereum will return the wallets to fund on Ethereum
func (keyStore *KeysStore) WalletsToFundOnEthereum() []common.Address {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([]common.Address, 0, len(allKeys))

	for _, key := range allKeys {
		if len(key.KlvSk) == 0 {
			continue
		}

		walletsToFund = append(walletsToFund, key.EthAddress)
	}

	return walletsToFund
}

// WalletsToFundOnKC will return the wallets to fund on KC
func (keyStore *KeysStore) WalletsToFundOnKC() []string {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([]string, 0, len(allKeys))

	for _, key := range allKeys {
		walletsToFund = append(walletsToFund, key.KlvAddress.Bech32())
	}

	return walletsToFund
}

// GenerateKlvPrivatePublicKey will generate a new keys holder instance that will hold only the KC generated keys
func GenerateKlvPrivatePublicKey(tb testing.TB, projectedShard byte) KeysHolder {
	sk, pkBytes := generateSkPkInShard(tb, projectedShard)

	skBytes, err := sk.ToByteArray()
	require.Nil(tb, err)

	return KeysHolder{
		KlvSk:      skBytes,
		KlvAddress: NewKlvAddressFromBytes(tb, pkBytes),
	}
}

func generateSkPkInShard(tb testing.TB, projectedShard byte) (klvCrypto.PrivateKey, []byte) {
	var sk klvCrypto.PrivateKey
	var pk klvCrypto.PublicKey

	for {
		sk, pk = keyGenerator.GeneratePair()

		pkBytes, err := pk.ToByteArray()
		require.Nil(tb, err)

		if pkBytes[len(pkBytes)-1] == projectedShard {
			return sk, pkBytes
		}
	}
}

// SaveKlvKey will save the Klever Blockchain key
func SaveKlvKey(tb testing.TB, filename string, key KeysHolder) {
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + key.KlvAddress.Bech32(),
		Bytes: []byte(hex.EncodeToString(key.KlvSk)),
	}

	buff := bytes.NewBuffer(make([]byte, 0))
	err := pem.Encode(buff, &blk)
	require.Nil(tb, err)

	err = os.WriteFile(filename, buff.Bytes(), os.ModePerm)
	require.Nil(tb, err)
}
