package sdk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/falcon-signatures/falcongo"
)

type AlgorandKeyInfo struct {
	AlgorandAddress string `json:"AlgorandAddress"`
	PublicKey       []byte `json:"PublicKey"`
	PrivateKey      []byte `json:"PrivateKey"`
}

const expectedMnemonicWords = 25

// MnemonicFromEntropy converts master-derivation-key entropy to its 25-word Algorand mnemonic.
func MnemonicFromEntropy(entropy []byte) (string, error) {
	var masterKey types.MasterDerivationKey
	if len(entropy) != len(masterKey) {
		return "", fmt.Errorf("invalid master derivation key length: got %d, want %d", len(entropy), len(masterKey))
	}
	copy(masterKey[:], entropy)
	return mnemonic.FromMasterDerivationKey(masterKey)
}

// MnemonicToEntropy converts an Algorand 25-word mnemonic to master-derivation-key entropy.
func MnemonicToEntropy(mnemonicStr string) ([]byte, error) {
	words := strings.Fields(strings.TrimSpace(mnemonicStr))
	if len(words) != expectedMnemonicWords {
		return nil, fmt.Errorf("mnemonic requires exactly %d words", expectedMnemonicWords)
	}

	entropy, err := mnemonic.ToKey(strings.Join(words, " "))
	if err != nil {
		return nil, fmt.Errorf("invalid Algorand mnemonic: %w", err)
	}
	return entropy, nil
}

// DeriveFromMnemonic derives a native Falcon-1024 PQ account from an Algorand 25-word mnemonic.
func DeriveFromMnemonic(mnemonicStr string, passphrase string) (*AlgorandKeyInfo, error) {
	seed, err := SeedFromMnemonic(mnemonicStr, passphrase)
	if err != nil {
		return nil, err
	}
	return keyInfoFromSeed(seed)
}

// SeedFromMnemonic derives the canonical Falcon-1024 PQ seed from an Algorand 25-word mnemonic.
// Algorand's PQ mnemonic derivation does not support passphrases.
func SeedFromMnemonic(mnemonicStr string, passphrase string) ([]byte, error) {
	if passphrase != "" {
		return nil, fmt.Errorf("passphrases are not supported for native Falcon-1024 accounts")
	}

	words := strings.Fields(strings.TrimSpace(mnemonicStr))
	if len(words) != expectedMnemonicWords {
		return nil, fmt.Errorf("mnemonic requires exactly %d words", expectedMnemonicWords)
	}

	seed, err := mnemonic.ToPQSeed(strings.Join(words, " "), types.PQSchemeFalcon1024)
	if err != nil {
		return nil, fmt.Errorf("invalid Algorand mnemonic: %w", err)
	}
	return seed, nil
}

// SeedFromEntropy derives the canonical Falcon-1024 PQ seed from master-key entropy.
// Algorand's PQ mnemonic derivation does not support passphrases.
func SeedFromEntropy(entropy []byte, passphrase string) ([]byte, error) {
	mnemonicStr, err := MnemonicFromEntropy(entropy)
	if err != nil {
		return nil, err
	}
	return SeedFromMnemonic(mnemonicStr, passphrase)
}

func keyInfoFromSeed(seed []byte) (*AlgorandKeyInfo, error) {
	account, err := crypto.Falcon1024AccountFromPQSeed(seed)
	if err != nil {
		return nil, err
	}
	return &AlgorandKeyInfo{
		AlgorandAddress: account.Address().String(),
		PublicKey:       account.PublicKey[:],
		PrivateKey:      account.PrivateKey[:],
	}, nil
}

// SignFalconBundle signs transactions from a native Falcon-1024 PQ account derived from seed.
// Transactions already signed by another account are preserved unchanged.
func SignFalconBundle(unsignedTxns *BytesArray, seed []byte) (string, error) {
	if unsignedTxns == nil || unsignedTxns.Length() == 0 {
		return "", fmt.Errorf("transaction bundle is empty")
	}

	account, err := crypto.Falcon1024AccountFromPQSeed(seed)
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
