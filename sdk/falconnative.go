package sdk

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"filippo.io/edwards25519"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	algorandMnemonic "github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/falcon-signatures/falcongo"
	"golang.org/x/crypto/pbkdf2"
)

type AlgorandKeyInfo struct {
	AlgorandAddress string `json:"AlgorandAddress"`
	PublicKey       []byte `json:"PublicKey"`
	PrivateKey      []byte `json:"PrivateKey"`
}

const (
	kdfIterations         = 100000
	kdfKeyLen             = 48
	kdfSaltStr            = "falcon-cli-seed-v1"
	expectedMnemonicWords = 25
)

var falcon1024Scheme = types.PQScheme{'f', '1'}

// MnemonicFromEntropy converts master-derivation-key entropy to its 25-word Algorand mnemonic.
func MnemonicFromEntropy(entropy []byte) (string, error) {
	var masterKey types.MasterDerivationKey
	if len(entropy) != len(masterKey) {
		return "", fmt.Errorf("invalid master derivation key length: got %d, want %d", len(entropy), len(masterKey))
	}
	copy(masterKey[:], entropy)
	return algorandMnemonic.FromMasterDerivationKey(masterKey)
}

// DeriveFromMnemonic derives a native Falcon-1024 PQ account from a 25-word mnemonic.
func DeriveFromMnemonic(mnemonicStr string, passphrase string) (*AlgorandKeyInfo, error) {
	words := strings.Fields(strings.TrimSpace(mnemonicStr))
	if len(words) != expectedMnemonicWords {
		return nil, fmt.Errorf("mnemonic requires exactly %d words", expectedMnemonicWords)
	}
	masterKey, err := algorandMnemonic.ToMasterDerivationKey(strings.Join(words, " "))
	if err != nil {
		return nil, fmt.Errorf("invalid Algorand mnemonic: %w", err)
	}
	seed := pbkdf2.Key(masterKey[:], []byte("falcon-native-account-v1"+passphrase), kdfIterations, kdfKeyLen, sha512.New)
	return keysFromSeed(seed)
}

// DeriveFromSeedPhrase deterministically derives a native Falcon-1024 PQ account.
func DeriveFromSeedPhrase(phrase string) (*AlgorandKeyInfo, error) {
	seed := pbkdf2.Key([]byte(strings.TrimSpace(phrase)), []byte(kdfSaltStr), kdfIterations, kdfKeyLen, sha512.New)
	return keysFromSeed(seed)
}

func keysFromSeed(seed []byte) (*AlgorandKeyInfo, error) {
	kp, err := falcongo.GenerateKeyPair(seed)
	if err != nil {
		return nil, err
	}
	_, address, err := canonicalPQAddress(kp.PublicKey[:])
	if err != nil {
		return nil, err
	}
	return &AlgorandKeyInfo{
		AlgorandAddress: address.String(),
		PublicKey:       kp.PublicKey[:],
		PrivateKey:      kp.PrivateKey[:],
	}, nil
}

// SignFalconBundle signs transactions from the native Falcon-1024 PQ account.
// Transactions already signed by another account are preserved unchanged.
func SignFalconBundle(unsignedTxns *BytesArray, pubKeyBytes []byte, privKeyBytes []byte) (string, error) {
	if unsignedTxns == nil || unsignedTxns.Length() == 0 {
		return "", fmt.Errorf("transaction bundle is empty")
	}

	salt, userAddress, err := canonicalPQAddress(pubKeyBytes)
	if err != nil {
		return "", err
	}

	keyPair := falcongo.KeyPair{}
	if len(pubKeyBytes) != len(keyPair.PublicKey) {
		return "", fmt.Errorf("invalid Falcon-1024 public key length: got %d, want %d", len(pubKeyBytes), len(keyPair.PublicKey))
	}
	if len(privKeyBytes) != len(keyPair.PrivateKey) {
		return "", fmt.Errorf("invalid Falcon-1024 private key length: got %d, want %d", len(privKeyBytes), len(keyPair.PrivateKey))
	}
	copy(keyPair.PublicKey[:], pubKeyBytes)
	copy(keyPair.PrivateKey[:], privKeyBytes)

	signedResults := make([]string, 0, unsignedTxns.Length())
	for i := 0; i < unsignedTxns.Length(); i++ {
		raw := unsignedTxns.Get(i)

		var existing types.SignedTxn
		if err := msgpack.Decode(raw, &existing); err == nil && isSigned(existing) {
			signedResults = append(signedResults, base64.StdEncoding.EncodeToString(raw))
			continue
		}

		var tx types.Transaction
		if err := msgpack.Decode(raw, &tx); err != nil {
			return "", fmt.Errorf("failed to decode transaction %d: %w", i, err)
		}
		if tx.Sender != userAddress {
			return "", fmt.Errorf("transaction %d sender %s is not the Falcon PQ account %s", i, tx.Sender, userAddress)
		}

		signature, err := keyPair.Sign(crypto.TransactionID(tx))
		if err != nil {
			return "", fmt.Errorf("sign transaction %d: %w", i, err)
		}
		stxn := types.SignedTxn{
			PQsig: types.PQSig{
				Scheme:    falcon1024Scheme,
				Salt:      salt,
				PublicKey: append([]byte(nil), pubKeyBytes...),
				Signature: append([]byte(nil), signature...),
			},
			Txn: tx,
		}
		signedResults = append(signedResults, base64.StdEncoding.EncodeToString(msgpack.Encode(&stxn)))
	}

	return strings.Join(signedResults, ","), nil
}

func isSigned(stxn types.SignedTxn) bool {
	return len(stxn.Sig) > 0 || len(stxn.Msig.Subsigs) > 0 || len(stxn.Lsig.Logic) > 0 || len(stxn.PQsig.Signature) > 0
}

func canonicalPQAddress(publicKey []byte) (types.PQAddressSalt, types.Address, error) {
	for salt := 0; salt <= 255; salt++ {
		addressHashInput := make([]byte, 0, len("PQA")+len(falcon1024Scheme)+1+len(publicKey))
		addressHashInput = append(addressHashInput, "PQA"...)
		addressHashInput = append(addressHashInput, falcon1024Scheme[:]...)
		addressHashInput = append(addressHashInput, byte(salt))
		addressHashInput = append(addressHashInput, publicKey...)
		addressHash := sha512.Sum512_256(addressHashInput)
		if _, err := new(edwards25519.Point).SetBytes(addressHash[:]); err != nil {
			var address types.Address
			copy(address[:], addressHash[:])
			return types.PQAddressSalt(salt), address, nil
		}
	}
	return 0, types.Address{}, fmt.Errorf("no canonical PQ address salt found")
}

func RawSign(messageBytes []byte, publicKeyBytes []byte, privateKeyBytes []byte) ([]byte, error) {
	keyPair := falcongo.KeyPair{}
	copy(keyPair.PublicKey[:], publicKeyBytes)
	copy(keyPair.PrivateKey[:], privateKeyBytes)
	return keyPair.Sign(messageBytes)
}

func (ki *AlgorandKeyInfo) ToJSON() (string, error) {
	data, err := json.MarshalIndent(ki, "", "  ")
	return string(data), err
}
