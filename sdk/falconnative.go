package sdk

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
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

// MnemonicFromEntropy converts master-derivation-key entropy to its 25-word Algorand mnemonic.
func MnemonicFromEntropy(entropy []byte) (string, error) {
	var masterKey types.MasterDerivationKey
	if len(entropy) != len(masterKey) {
		return "", fmt.Errorf("invalid master derivation key length: got %d, want %d", len(entropy), len(masterKey))
	}
	copy(masterKey[:], entropy)
	return mnemonic.FromMasterDerivationKey(masterKey)
}

// DeriveFromMnemonic derives a native Falcon-1024 PQ account from a 25-word mnemonic.
func DeriveFromMnemonic(mnemonicStr string, passphrase string) (*AlgorandKeyInfo, error) {
	words := strings.Fields(strings.TrimSpace(mnemonicStr))
	if len(words) != expectedMnemonicWords {
		return nil, fmt.Errorf("mnemonic requires exactly %d words", expectedMnemonicWords)
	}
	masterKey, err := mnemonic.ToMasterDerivationKey(strings.Join(words, " "))
	if err != nil {
		return nil, fmt.Errorf("invalid Algorand mnemonic: %w", err)
	}
	account, err := falconAccountFromEntropy(masterKey[:], passphrase)
	if err != nil {
		return nil, err
	}
	return &AlgorandKeyInfo{
		AlgorandAddress: account.Address().String(),
		PublicKey:       account.PublicKey[:],
		PrivateKey:      account.PrivateKey[:],
	}, nil
}

func falconAccountFromEntropy(entropy []byte, passphrase string) (crypto.Falcon1024Account, error) {
	var masterKey types.MasterDerivationKey
	if len(entropy) != len(masterKey) {
		return crypto.Falcon1024Account{}, fmt.Errorf("invalid master derivation key length: got %d, want %d", len(entropy), len(masterKey))
	}
	seed := pbkdf2.Key(entropy, []byte("falcon-native-account-v1"+passphrase), kdfIterations, kdfKeyLen, sha512.New)
	return crypto.Falcon1024AccountFromPQSeed(seed)
}

// SignFalconBundle signs transactions from a native Falcon-1024 PQ account derived from master-key entropy.
// Transactions already signed by another account are preserved unchanged.
func SignFalconBundle(unsignedTxns *BytesArray, entropy []byte, passphrase string) (string, error) {
	if unsignedTxns == nil || unsignedTxns.Length() == 0 {
		return "", fmt.Errorf("transaction bundle is empty")
	}

	account, err := falconAccountFromEntropy(entropy, passphrase)
	if err != nil {
		return "", err
	}
	userAddress := account.Address()

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

		_, signedTxn, err := crypto.SignFalcon1024AccountTransaction(account, tx)
		if err != nil {
			return "", fmt.Errorf("sign transaction %d: %w", i, err)
		}
		signedResults = append(signedResults, base64.StdEncoding.EncodeToString(signedTxn))
	}

	return strings.Join(signedResults, ","), nil
}

func isSigned(stxn types.SignedTxn) bool {
	return len(stxn.Sig) > 0 || len(stxn.Msig.Subsigs) > 0 || len(stxn.Lsig.Logic) > 0 || len(stxn.PQsig.Signature) > 0
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
