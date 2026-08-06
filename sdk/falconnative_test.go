package sdk

import (
	"encoding/base64"
	"testing"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleAlgorandMnemonic = "carbon another pair valley ride lumber exhibit chunk forget select nerve topic refuse ball bomb draw chunk toward motor detect process smile envelope abstract rule"

func TestMnemonicFromEntropy(t *testing.T) {
	masterKey, err := mnemonic.ToMasterDerivationKey(sampleAlgorandMnemonic)
	require.NoError(t, err)

	mnemonic, err := MnemonicFromEntropy(masterKey[:])
	require.NoError(t, err)
	assert.Equal(t, sampleAlgorandMnemonic, mnemonic)

	entropy, err := MnemonicToEntropy("  " + sampleAlgorandMnemonic + "  ")
	require.NoError(t, err)
	assert.Equal(t, masterKey[:], entropy)

	_, err = MnemonicToEntropy("invalid mnemonic")
	assert.Error(t, err)

	_, err = MnemonicFromEntropy(masterKey[:len(masterKey)-1])
	assert.Error(t, err)

	seed, err := SeedFromEntropy(masterKey[:], "")
	require.NoError(t, err)
	assert.Len(t, seed, 32)

	_, err = SeedFromEntropy(masterKey[:], "test passphrase")
	assert.Error(t, err)

	_, err = SeedFromEntropy(masterKey[:len(masterKey)-1], "")
	assert.Error(t, err)
}

func TestSeedFromMnemonicUsesAlgorandSDK(t *testing.T) {
	seed, err := SeedFromMnemonic(sampleAlgorandMnemonic, "")
	require.NoError(t, err)

	expectedSeed, err := mnemonic.ToPQSeed(sampleAlgorandMnemonic, types.PQSchemeFalcon1024)
	require.NoError(t, err)
	assert.Equal(t, expectedSeed, seed)

	_, err = SeedFromMnemonic(sampleAlgorandMnemonic, "passphrase")
	assert.Error(t, err)
}

func TestFalconNativeAccountDerivation(t *testing.T) {
	keyInfo, err := DeriveFromMnemonic(sampleAlgorandMnemonic, "")
	require.NoError(t, err)

	var expectedAccount crypto.Falcon1024Account
	assert.Len(t, keyInfo.PublicKey, len(expectedAccount.PublicKey), "Falcon-1024 public key length")
	assert.Len(t, keyInfo.PrivateKey, len(expectedAccount.PrivateKey), "Falcon-1024 private key length")

	masterKey, err := mnemonic.ToMasterDerivationKey(sampleAlgorandMnemonic)
	require.NoError(t, err)
	seed, err := SeedFromEntropy(masterKey[:], "")
	require.NoError(t, err)
	account, err := crypto.Falcon1024AccountFromPQSeed(seed)
	require.NoError(t, err)
	assert.Equal(t, account.Address().String(), keyInfo.AlgorandAddress)
}

func TestSignFalconBundleCreatesNativePQSignature(t *testing.T) {
	keyInfo, err := DeriveFromMnemonic(sampleAlgorandMnemonic, "")
	require.NoError(t, err)

	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}
	amount := MakeUint64(0)
	unsignedTxn, err := MakePaymentTxn(keyInfo.AlgorandAddress, keyInfo.AlgorandAddress, &amount, nil, "", params)
	require.NoError(t, err)

	masterKey, err := mnemonic.ToMasterDerivationKey(sampleAlgorandMnemonic)
	require.NoError(t, err)
	seed, err := SeedFromEntropy(masterKey[:], "")
	require.NoError(t, err)
	txnList := &BytesArray{}
	txnList.Append(unsignedTxn)
	_, err = SignFalconBundle(txnList, seed[:len(seed)-1])
	require.Error(t, err)

	signedBundle, err := SignFalconBundle(txnList, seed)
	require.NoError(t, err)

	encodedNativeTxn, err := base64.StdEncoding.DecodeString(signedBundle)
	require.NoError(t, err)
	var nativeTxn types.SignedTxn
	require.NoError(t, msgpack.Decode(encodedNativeTxn, &nativeTxn))
	assert.Equal(t, types.PQSchemeFalcon1024, nativeTxn.PQsig.Scheme)
	assert.Equal(t, keyInfo.PublicKey, nativeTxn.PQsig.PublicKey)
	assert.Empty(t, nativeTxn.Lsig.Logic)
	assert.NotEmpty(t, nativeTxn.PQsig.Signature)

	toBeSigned := append([]byte("TX"), msgpack.Encode(nativeTxn.Txn)...)
	assert.True(t, crypto.VerifyPQSig(toBeSigned, nativeTxn.PQsig))
}

func mustDecodeBase64(t *testing.T, s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return data
}
