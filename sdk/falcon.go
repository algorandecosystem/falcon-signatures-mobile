package sdk

import (
	_ "embed"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/falcon-signatures/algorand"
	"github.com/algorandfoundation/falcon-signatures/falcongo"
	"github.com/algorandfoundation/falcon-signatures/mnemonic"
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
	expectedMnemonicWords = 24
	dummiesPerRealTxn     = 3
)

//go:embed teal/dummyLsig.teal.tok
var dummyLsigCompiled []byte

// --- Key Management ---

func DeriveFromMnemonic(mnemonicStr string, passphrase string) (*AlgorandKeyInfo, error) {
	words := strings.Fields(strings.TrimSpace(mnemonicStr))
	if len(words) != expectedMnemonicWords {
		return nil, fmt.Errorf("mnemonic requires exactly %d words", expectedMnemonicWords)
	}
	seedArray, err := mnemonic.SeedFromMnemonic(words, passphrase)
	if err != nil {
		return nil, err
	}
	return keysFromSeed(seedArray[:])
}

func DeriveFromSeedPhrase(phrase string) (*AlgorandKeyInfo, error) {
	seed := pbkdf2.Key([]byte(strings.TrimSpace(phrase)), []byte(kdfSaltStr), kdfIterations, kdfKeyLen, sha512.New)
	return keysFromSeed(seed)
}

func keysFromSeed(seed []byte) (*AlgorandKeyInfo, error) {
	kp, err := falcongo.GenerateKeyPair(seed)
	if err != nil {
		return nil, err
	}
	address, err := algorand.GetAddressFromPublicKey(kp.PublicKey)
	if err != nil {
		return nil, err
	}
	return &AlgorandKeyInfo{
		AlgorandAddress: string(address),
		PublicKey:       kp.PublicKey[:],
		PrivateKey:      kp.PrivateKey[:],
	}, nil
}

// --- Signing Logic ---

// SignFalconBundle handles multiple transactions (raw bytes).
// Returns a comma-separated string of signed Base64 transactions.
//
// TWO MODES:
// 1. No group ID (single txn): Add dummies, create group, sign
// 2. Has group ID (from dApp): Just sign as-is, don't modify!
func SignFalconBundle(
	unsignedTxns *BytesArray,
	pubKeyBytes []byte,
	privKeyBytes []byte,
) (string, error) {
	var txns []types.Transaction
	var alreadySigned []bool
	var originalBytes [][]byte // Store original signed bytes to preserve signatures

	// 1. Decode all incoming transactions
	for i := 0; i < unsignedTxns.Length(); i++ {
		raw := unsignedTxns.Get(i)

		// Try to decode as SignedTxn first (check if already signed)
		var stxn types.SignedTxn
		if err := msgpack.Decode(raw, &stxn); err == nil && (len(stxn.Sig) > 0 || len(stxn.Lsig.Logic) > 0) {
			txns = append(txns, stxn.Txn)
			alreadySigned = append(alreadySigned, true)
			originalBytes = append(originalBytes, raw) // Save original bytes to preserve signature
			continue
		}

		// Not signed - decode as unsigned Transaction
		var t types.Transaction
		if err := msgpack.Decode(raw, &t); err != nil {
			return "", fmt.Errorf("failed to decode transaction %d: %w", i, err)
		}
		txns = append(txns, t)
		alreadySigned = append(alreadySigned, false)
		originalBytes = append(originalBytes, nil) // No original bytes for unsigned txns
	}

	// 2. CHECK: Does the first txn have a group ID?
	// If NO group ID: This is a single txn call - add dummies and create group
	// If HAS group ID: This is from dApp - DON'T MODIFY, just sign!
	if len(txns) > 0 && txns[0].Group == (types.Digest{}) {
		// MODE 1: No group ID - add dummies for budget
		fmt.Println("SignFalconBundle: No group ID detected - adding dummies")

        actualDummyCount := len(txns) * dummiesPerRealTxn  // 3 dummies per txn

        fmt.Printf("SignFalconBundle: Adding %d dummies for %d real txns (total: %d)\n",
            actualDummyCount, len(txns), len(txns)+actualDummyCount)

		if actualDummyCount > 0 {
			for i := 0; i < actualDummyCount; i++ {
				txns = append(txns, createDummyTransaction(txns[0], i))
				alreadySigned = append(alreadySigned, false)
			}
			// Add fees to first txn to cover dummies
			txns[0].Fee += types.MicroAlgos(uint64(actualDummyCount) * 1000)
		}

		// Compute and apply group ID
		gid, err := crypto.ComputeGroupID(txns)
		if err != nil {
			return "", fmt.Errorf("failed to compute group ID: %w", err)
		}
		for i := range txns {
			txns[i].Group = gid
		}
	} else {
		// MODE 2: Has group ID - DON'T MODIFY ANYTHING!
		fmt.Println("SignFalconBundle: Group ID detected - signing as-is")
		// Don't add dummies, don't modify fees, don't recompute group!
	}

	// 3. Sign all transactions
	var signedResults []string
	keyPair := falcongo.KeyPair{}
	copy(keyPair.PublicKey[:], pubKeyBytes)
	copy(keyPair.PrivateKey[:], privKeyBytes)

	userAddressStr, _ := algorand.GetAddressFromPublicKey(keyPair.PublicKey)
	userAddress, _ := types.DecodeAddress(string(userAddressStr))

	for i := range txns {
		var encoded []byte

		if alreadySigned[i] {
			// Already signed - use original bytes to preserve signature
			encoded = originalBytes[i]
		} else if txns[i].Sender == userAddress {
			// User txn: Falcon LogicSig
			lsig, _ := algorand.DerivePQLogicSig(keyPair.PublicKey)
			sig, _ := keyPair.Sign(crypto.TransactionID(txns[i]))
			lsig.Lsig.Args = [][]byte{sig}
			stxn := types.SignedTxn{Lsig: lsig.Lsig, Txn: txns[i]}
			encoded = msgpack.Encode(&stxn)
		} else {
			// Dummy: Standard LogicSig
			lsig := types.LogicSig{Logic: dummyLsigCompiled}
			_, signedBytes, _ := crypto.SignLogicSigTransaction(lsig, txns[i])
			encoded = signedBytes
		}
		signedResults = append(signedResults, base64.StdEncoding.EncodeToString(encoded))
	}

	return strings.Join(signedResults, ","), nil
}

// --- Helpers ---

func createDummyTransaction(template types.Transaction, index int) types.Transaction {
	lsig := crypto.LogicSigAccount{Lsig: types.LogicSig{Logic: dummyLsigCompiled}}
	addr, _ := lsig.Address()

	return types.Transaction{
		Type: types.PaymentTx,
		Header: types.Header{
			Sender:      types.Address(addr),
			Fee:         0,
			FirstValid:  template.FirstValid,
			LastValid:   template.LastValid,
			GenesisHash: template.GenesisHash,
			GenesisID:   template.GenesisID,
			Note:        []byte{byte(index)},
		},
		PaymentTxnFields: types.PaymentTxnFields{
			Receiver: types.Address(addr),
			Amount:   0,
		},
	}
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
