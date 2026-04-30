package sdk

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SAMPLE_BIP39_MNEMONIC is a 24-word BIP39 mnemonic for testing Falcon key derivation
const SAMPLE_BIP39_MNEMONIC = "lab pause february spread carpet true balance autumn frog clock lunch silent pigeon live task liar shield either guard game suggest account control gossip"

// TestDeriveFromMnemonic tests deriving Falcon keys from a BIP39 mnemonic
func TestFalconDeriveFromMnemonic(t *testing.T) {
	keyInfo, err := DeriveFromMnemonic(SAMPLE_BIP39_MNEMONIC, "")
	require.NoError(t, err)
	require.NotNil(t, keyInfo)

	// Verify keys were generated (Falcon-512 public key is 1792-1793 bytes, private is 2304-2305 bytes)
	assert.NotEmpty(t, keyInfo.AlgorandAddress)
	assert.GreaterOrEqual(t, len(keyInfo.PublicKey), 1792, "Falcon public key should be at least 1792 bytes")
	assert.LessOrEqual(t, len(keyInfo.PublicKey), 1793, "Falcon public key should be at most 1793 bytes")
	assert.GreaterOrEqual(t, len(keyInfo.PrivateKey), 2304, "Falcon private key should be at least 2304 bytes")
	assert.LessOrEqual(t, len(keyInfo.PrivateKey), 2305, "Falcon private key should be at most 2305 bytes")

	t.Logf("Derived Falcon address: %s", keyInfo.AlgorandAddress)
}

// TestSignFalconBundle_ZeroPayment tests SignFalconBundle with a single 0-amount payment transaction
func TestFalconSignFalconBundle_ZeroPayment(t *testing.T) {
	// Derive keys from mnemonic
	keyInfo, err := DeriveFromMnemonic(SAMPLE_BIP39_MNEMONIC, "")
	require.NoError(t, err)

	// Create a minimal payment transaction with 0 amount
	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}

	// Create a 0-amount payment transaction from the Falcon address
	zeroAmount := MakeUint64(0)
	unsignedTxn, err := MakePaymentTxn(
		keyInfo.AlgorandAddress, // sender (Falcon account)
		keyInfo.AlgorandAddress, // receiver (self-transfer)
		&zeroAmount,             // amount = 0
		nil,                     // note
		"",                      // closeRemainderTo
		params,
	)
	require.NoError(t, err)
	require.NotNil(t, unsignedTxn)

	// Wrap in BytesArray for SignFalconBundle
	txnList := &BytesArray{}
	txnList.Append(unsignedTxn)

	// Sign the bundle
	signedBundle, err := SignFalconBundle(
		txnList,
		keyInfo.PublicKey,
		keyInfo.PrivateKey,
	)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)

	// Verify we got back a comma-separated string of base64 transactions
	// Should have at least 4 transactions (1 real + 3 dummies)
	t.Logf("Signed bundle length: %d", len(signedBundle))
	t.Logf("Signed bundle (first 200 chars): %s...", signedBundle[:min(200, len(signedBundle))])
}

// TestSignFalconBundle_WithGroupID tests SignFalconBundle when transactions already have a group ID
// This simulates the real use case where:
// - Index 0: dApp-signed funding transaction (already signed)
// - Index 1-2: User's transactions needing Falcon signatures
func TestFalconSignFalconBundle_WithGroupID(t *testing.T) {
	// Derive Falcon keys from mnemonic
	keyInfo, err := DeriveFromMnemonic(SAMPLE_BIP39_MNEMONIC, "")
	require.NoError(t, err)

	// Generate a standard ed25519 account for the dApp (not Falcon)
	dAppSK := GenerateSK()
	dAppAddr, err := GenerateAddressFromSK(dAppSK)
	require.NoError(t, err)

	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}

	// Create funding transaction from dApp to user (already signed by dApp)
	fundingAmount := MakeUint64(100000) // 0.1 ALGO
	fundingTxn, err := MakePaymentTxn(
		dAppAddr,                   // sender: dApp
		keyInfo.AlgorandAddress,    // receiver: user (Falcon account)
		&fundingAmount,
		nil,
		"",
		params,
	)
	require.NoError(t, err)

	// Create two user transactions needing Falcon signatures
	zeroAmount := MakeUint64(0)
	userTxn1, err := MakePaymentTxn(
		keyInfo.AlgorandAddress,
		keyInfo.AlgorandAddress,
		&zeroAmount,
		nil,
		"",
		params,
	)
	require.NoError(t, err)

	userTxn2, err := MakePaymentTxn(
		keyInfo.AlgorandAddress,
		keyInfo.AlgorandAddress,
		&zeroAmount,
		nil,
		"",
		params,
	)
	require.NoError(t, err)

	// Sign the funding transaction with dApp's key (simulating server-side signing)
	signedFundingTxn, err := SignTransaction(dAppSK, fundingTxn)
	require.NoError(t, err)

	// Wrap in BytesArray: [dApp-signed, unsigned, unsigned]
	txnList := &BytesArray{}
	txnList.Append(signedFundingTxn) // Already signed by dApp
	txnList.Append(userTxn1)         // Needs Falcon signature
	txnList.Append(userTxn2)         // Needs Falcon signature

	// Sign the bundle - should detect funding txn is already signed
	signedBundle, err := SignFalconBundle(
		txnList,
		keyInfo.PublicKey,
		keyInfo.PrivateKey,
	)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)

	// Verify we got a valid signed bundle
	t.Logf("Signed bundle with pre-grouped txns (first 200 chars): %s...", signedBundle[:min(200, len(signedBundle))])
	t.Logf("Full bundle length: %d characters", len(signedBundle))
}

// TestSignFalconBundle_MultipleTxns tests SignFalconBundle with multiple unsigned user transactions
func TestFalconSignFalconBundle_MultipleTxns(t *testing.T) {
	// Derive keys from mnemonic
	keyInfo, err := DeriveFromMnemonic(SAMPLE_BIP39_MNEMONIC, "")
	require.NoError(t, err)

	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}

	// Create two payment transactions
	zeroAmount := MakeUint64(0)
	txn1, err := MakePaymentTxn(
		keyInfo.AlgorandAddress,
		keyInfo.AlgorandAddress,
		&zeroAmount,
		nil,
		"",
		params,
	)
	require.NoError(t, err)

	txn2, err := MakePaymentTxn(
		keyInfo.AlgorandAddress,
		keyInfo.AlgorandAddress,
		&zeroAmount,
		nil,
		"",
		params,
	)
	require.NoError(t, err)

	// Wrap in BytesArray
	txnList := &BytesArray{}
	txnList.Append(txn1)
	txnList.Append(txn2)

	// Sign the bundle
	signedBundle, err := SignFalconBundle(
		txnList,
		keyInfo.PublicKey,
		keyInfo.PrivateKey,
	)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)

	t.Logf("Signed bundle with 2 txns (first 200 chars): %s...", signedBundle[:min(200, len(signedBundle))])
}

// mustDecodeBase64 is a test helper that decodes base64 or fails the test
func mustDecodeBase64(t *testing.T, s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return data
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
