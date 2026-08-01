package sdk

import (
	"testing"

	"github.com/algorandfoundation/falcon-signatures/algorand"
	"github.com/algorandfoundation/falcon-signatures/falcongo"
	"github.com/stretchr/testify/require"
)

const sampleBIP39Mnemonic = "lab pause february spread carpet true balance autumn frog clock lunch silent pigeon live task liar shield either guard game suggest account control gossip"

func TestSignFalconLsigBundleZeroPayment(t *testing.T) {
	keyInfo, err := DeriveFalconLsigFromMnemonic(sampleBIP39Mnemonic, "")
	require.NoError(t, err)
	legacyAddress, err := algorand.GetAddressFromPublicKey(toFalconPublicKey(t, keyInfo.PublicKey))
	require.NoError(t, err)

	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}
	amount := MakeUint64(0)
	unsignedTxn, err := MakePaymentTxn(string(legacyAddress), string(legacyAddress), &amount, nil, "", params)
	require.NoError(t, err)

	txnList := &BytesArray{}
	txnList.Append(unsignedTxn)
	signedBundle, err := SignFalconLsigBundle(txnList, keyInfo.PublicKey, keyInfo.PrivateKey)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)
}

func toFalconPublicKey(t *testing.T, publicKeyBytes []byte) falcongo.PublicKey {
	t.Helper()
	var publicKey falcongo.PublicKey
	require.Len(t, publicKeyBytes, len(publicKey))
	copy(publicKey[:], publicKeyBytes)
	return publicKey
}

func TestSignFalconLsigBundleWithGroupID(t *testing.T) {
	keyInfo, err := DeriveFalconLsigFromMnemonic(sampleBIP39Mnemonic, "")
	require.NoError(t, err)
	legacyAddress, err := algorand.GetAddressFromPublicKey(toFalconPublicKey(t, keyInfo.PublicKey))
	require.NoError(t, err)

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

	fundingAmount := MakeUint64(100000)
	fundingTxn, err := MakePaymentTxn(dAppAddr, string(legacyAddress), &fundingAmount, nil, "", params)
	require.NoError(t, err)
	amount := MakeUint64(0)
	userTxn1, err := MakePaymentTxn(string(legacyAddress), string(legacyAddress), &amount, nil, "", params)
	require.NoError(t, err)
	userTxn2, err := MakePaymentTxn(string(legacyAddress), string(legacyAddress), &amount, nil, "", params)
	require.NoError(t, err)
	signedFundingTxn, err := SignTransaction(dAppSK, fundingTxn)
	require.NoError(t, err)

	txnList := &BytesArray{}
	txnList.Append(signedFundingTxn)
	txnList.Append(userTxn1)
	txnList.Append(userTxn2)
	signedBundle, err := SignFalconLsigBundle(txnList, keyInfo.PublicKey, keyInfo.PrivateKey)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)
}

func TestSignFalconLsigBundleMultipleTransactions(t *testing.T) {
	keyInfo, err := DeriveFalconLsigFromMnemonic(sampleBIP39Mnemonic, "")
	require.NoError(t, err)
	legacyAddress, err := algorand.GetAddressFromPublicKey(toFalconPublicKey(t, keyInfo.PublicKey))
	require.NoError(t, err)

	params := &SuggestedParams{
		GenesisID:       "testnet-v1.0",
		GenesisHash:     mustDecodeBase64(t, "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiA="),
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		Fee:             1000,
	}
	amount := MakeUint64(0)
	txn1, err := MakePaymentTxn(string(legacyAddress), string(legacyAddress), &amount, nil, "", params)
	require.NoError(t, err)
	txn2, err := MakePaymentTxn(string(legacyAddress), string(legacyAddress), &amount, nil, "", params)
	require.NoError(t, err)

	txnList := &BytesArray{}
	txnList.Append(txn1)
	txnList.Append(txn2)
	signedBundle, err := SignFalconLsigBundle(txnList, keyInfo.PublicKey, keyInfo.PrivateKey)
	require.NoError(t, err)
	require.NotEmpty(t, signedBundle)
}
