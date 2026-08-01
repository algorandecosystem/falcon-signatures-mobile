package sdk

import (
	"encoding/base64"
	"testing"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/falcon-signatures/falcongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleAlgorandMnemonic = "carbon another pair valley ride lumber exhibit chunk forget select nerve topic refuse ball bomb draw chunk toward motor detect process smile envelope abstract rule"

func TestFalconNativeAccountDerivation(t *testing.T) {
	keyInfo, err := DeriveFromMnemonic(sampleAlgorandMnemonic, "")
	require.NoError(t, err)

	var expectedKeyPair falcongo.KeyPair
	assert.Len(t, keyInfo.PublicKey, len(expectedKeyPair.PublicKey), "Falcon-1024 public key length")
	assert.Len(t, keyInfo.PrivateKey, len(expectedKeyPair.PrivateKey), "Falcon-1024 private key length")

	salt, derivedAddress, err := canonicalPQAddress(keyInfo.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, derivedAddress.String(), keyInfo.AlgorandAddress)
	assert.LessOrEqual(t, uint8(salt), uint8(255))
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

	txnList := &BytesArray{}
	txnList.Append(unsignedTxn)
	signedBundle, err := SignFalconBundle(txnList, keyInfo.PublicKey, keyInfo.PrivateKey)
	require.NoError(t, err)

	encodedNativeTxn, err := base64.StdEncoding.DecodeString(signedBundle)
	require.NoError(t, err)
	var nativeTxn types.SignedTxn
	require.NoError(t, msgpack.Decode(encodedNativeTxn, &nativeTxn))
	assert.Equal(t, falcon1024Scheme, nativeTxn.PQsig.Scheme)
	assert.Equal(t, keyInfo.PublicKey, nativeTxn.PQsig.PublicKey)
	assert.Empty(t, nativeTxn.Lsig.Logic)
	assert.NotEmpty(t, nativeTxn.PQsig.Signature)

	var publicKey falcongo.PublicKey
	copy(publicKey[:], nativeTxn.PQsig.PublicKey)
	assert.NoError(t, falcongo.Verify(crypto.TransactionID(nativeTxn.Txn), nativeTxn.PQsig.Signature, publicKey))
}

func mustDecodeBase64(t *testing.T, s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return data
}
