package cubist

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	bitcoinlib "github.com/axengine/gogolibs/bitcoin"
	"github.com/axengine/utils"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var _cli_ *Cubist

func TestMain(m *testing.M) {
	_cli_ = New(true, "")
	os.Exit(m.Run())
}

func TestMe(t *testing.T) {
	rsp, err := _cli_.Me()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestOrg(t *testing.T) {
	rsp, err := _cli_.Org()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestKeys(t *testing.T) {
	rsp, err := _cli_.Keys()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestCreateSession(t *testing.T) {
	rsp, err := _cli_.createSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))

	time.Sleep(time.Minute)
	if err := _cli_.refreshToken(rsp); err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestRefreshSession(t *testing.T) {
	s := `{
          "org_id": "Org#847b77bc-1cc0-4ad3-9b70-f991c3e8b699",
          "role_id": "",
          "expiration": 1774580623,
          "purpose": "cs v0.85.0+fe242997",
          "token": "3d6fd7397:MDhmYzU5YTktYTgwNC00MGFiLThjMWEtYmVjNmVkYTBhOWY2.eyJlcG9jaF9udW0iOjEsImVwb2NoX3Rva2VuIjoiOXpudVBBSnU5UGhQbU5udUtNdk0zdG9zWWVDWXZITXdZVGcycnFLamoyYz0iLCJvdGhlcl90b2tlbiI6InBYRk1xdk5iSFdjWlM3Mm8vd09ETTBlMGVCcWFZbGk2VXVhWW1QNXg3c009In0=",
          "refresh_token": "3d6fd7397:MDhmYzU5YTktYTgwNC00MGFiLThjMWEtYmVjNmVkYTBhOWY2.eyJlcG9jaF9udW0iOjEsImVwb2NoX3Rva2VuIjoiOXpudVBBSnU5UGhQbU5udUtNdk0zdG9zWWVDWXZITXdZVGcycnFLamoyYz0iLCJvdGhlcl90b2tlbiI6InBYRk1xdk5iSFdjWlM3Mm8vd09ETTBlMGVCcWFZbGk2VXVhWW1QNXg3c009In0=.Tboc7pRzu52iKJC/0Tb8lwUmNUU4yy/59n4Vga9i8Ak=",
          "env": {
            "Dev-CubeSignerStack": {
              "ClientId": "1tiou9ecj058khiidmhj4ds4rj",
              "GoogleDeviceClientId": "59575607964-nc9hjnjka7jlb838jmg40qes4dtpsm6e.apps.googleusercontent.com",
              "GoogleDeviceClientSecret": "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",
              "Region": "us-east-1",
              "UserPoolId": "us-east-1_RU7HEslOW",
              "SignerApiRoot": "https://gamma.signer.cubist.dev",
              "DefaultCredentialRpId": "cubist.dev",
              "EncExportS3BucketName": null,
              "DeletedKeysS3BucketName": null
            }
          },
          "session_info": {
            "auth_token": "pXFMqvNbHWcZS72o/wODM0e0eBqaYli6UuaYmP5x7sM=",
            "auth_token_exp": 1743044946,
            "epoch": 1,
            "epoch_token": "9znuPAJu9PhPmNnuKMvM3tosYeCYvHMwYTg2rqKjj2c=",
            "refresh_token": "Tboc7pRzu52iKJC/0Tb8lwUmNUU4yy/59n4Vga9i8Ak=",
            "refresh_token_exp": 1743131046,
            "session_id": "08fc59a9-a804-40ab-8c1a-bec6eda0a9f6"
          }
        }`

	var session Session
	json.Unmarshal([]byte(s), &session)
	if err := _cli_.refreshToken(&session); err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(session))
}

func TestCubist_PsbtSign(t *testing.T) {
	rsp, err := _cli_.PsbtSign(context.Background(),
		"tb1qtwdzg7zyrnvf2jhlatydsx68svqzswannlzyzw",
		"70736274ff01005e020000000162cf792c5687a311f369a1b489efa62393d1897a2c16111c25f8d26c8670e221410000000000fdffffff01905f010000000000225120b9e42233471b8d58e9cb98fda7ae088f16ff6fe71733c061ce73ea96ec613cd6000000000001012ba08601000000000022002095a954ca6b8050d45510dda16d19698b599cc46d27626ac359c840a6ba4f7624010569522103ce80660233a78c36cf98e9957566d67406aa094daa5c1dfeee682a8334861da42102b7b38122d8507d907c53f0c60b099e9ecd94210fd5b4f6cf9cb9989c778a9674210287fd11b345a80b38a6ae5a8fbe4e31fbe344e83bb9ad120d7889b4f00587efd753ae0000",
		nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func encodeToBase64Url(buffer []byte) string {
	// Encode the buffer to standard Base64
	b64 := base64.StdEncoding.EncodeToString(buffer)
	// Replace URL-unsafe characters with URL-safe ones and remove padding
	b64 = strings.ReplaceAll(b64, "+", "-")
	b64 = strings.ReplaceAll(b64, "/", "_")
	b64 = strings.TrimRight(b64, "=")
	return b64
}

func TestCubist_PsbtSignWithMFA(t *testing.T) {
	headers := make(map[string]string)

	receipt := make(map[string]interface{})
	receipt["id"] = "MfaRequest#1d6ac5b5-365a-4346-90dd-835525a5e483"
	receipt["confirmation"] = "710f9633-38dd-48f9-96d9-ceb5ad73398d"

	var receipts []map[string]interface{}
	receipts = append(receipts, receipt)
	bz, _ := json.Marshal(receipts)
	headers["x-cubist-mfa-org-id"] = "Org#847b77bc-1cc0-4ad3-9b70-f991c3e8b699"
	headers["x-cubist-mfa-receipts"] = encodeToBase64Url(bz)

	rsp, err := _cli_.PsbtSign(context.Background(),
		"tb1qtwdzg7zyrnvf2jhlatydsx68svqzswannlzyzw",
		"70736274ff01005e020000000162cf792c5687a3f369a1b489efa62393d1897a2c16111c25f8d26c8670e221410000000000fdffffff01905f010000000000225120b9e42233471b8d58e9cb98fda7ae088f16ff6fe71733c061ce73ea96ec613cd6000000000001012ba08601000000000022002095a954ca6b8050d45510dda16d19698b599cc46d27626ac359c840a6ba4f7624010569522103ce80660233a78c36cf98e9957566d67406aa094daa5c1dfeee682a8334861da42102b7b38122d8507d907c53f0c60b099e9ecd94210fd5b4f6cf9cb9989c778a9674210287fd11b345a80b38a6ae5a8fbe4e31fbe344e83bb9ad120d7889b4f00587efd753ae0000",
		headers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestCubist_EIP191Sign(t *testing.T) {
	// This permission is linked to the key. Reference: https://signer-docs.cubist.dev/scopes-policies-roles/sign-manage-policies.html
	// cs key set-policy --key-id "Key#0xa1ec06ef11eb47f98bd919b5bf27c2107e44fd9f" --policy '"AllowEip191Signing"
	rsp, err := _cli_.EIP191Sign(context.Background(), "0xa1ec06ef11eb47f98bd919b5bf27c2107e44fd9f", "0x1234567890", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestCubist_EIP712Sign(t *testing.T) {
	// cs key set-policy --key-id "Key#0xa1ec06ef11eb47f98bd919b5bf27c2107e44fd9f" --policy '"AllowEip712Signing"

	typesSafeTx := Types{
		"EIP712Domain": {
			//{
			//	Name: "name",
			//	Type: "string",
			//},
			//{
			//	Name: "version",
			//	Type: "string",
			//},
			{
				Name: "chainId",
				Type: "uint256",
			},
			{
				Name: "verifyingContract",
				Type: "address",
			},
			//{
			//	Name: "salt",
			//	Type: "string",
			//},
		},
		"SafeTx": {
			{
				Name: "to",
				Type: "address",
			},
			{
				Name: "value",
				Type: "uint256",
			},
			{
				Name: "data",
				Type: "bytes",
			},
			{
				Name: "operation",
				Type: "uint8",
			},
			{
				Name: "safeTxGas",
				Type: "uint256",
			},
			{
				Name: "baseGas",
				Type: "uint256",
			},
			{
				Name: "gasPrice",
				Type: "uint256",
			},
			{
				Name: "gasToken",
				Type: "address",
			},
			{
				Name: "refundReceiver",
				Type: "address",
			},
			{
				Name: "nonce",
				Type: "uint256",
			},
		},
	}

	data := &TypedData{
		Types:       typesSafeTx,
		PrimaryType: "SafeTx",
		Domain: TypedDataDomain{
			//Name:              "",
			//Version:           "",
			ChainId:           math.NewHexOrDecimal256(1337),
			VerifyingContract: "0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
			//Salt:              "",
		},
		Message: apitypes.TypedDataMessage{
			"to":             common.HexToAddress("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC"),
			"value":          math.NewHexOrDecimal256(1),
			"data":           hexutil.Encode([]byte("hello")),
			"operation":      math.NewHexOrDecimal256(0),
			"safeTxGas":      math.NewHexOrDecimal256(1),
			"baseGas":        math.NewHexOrDecimal256(1),
			"gasPrice":       math.NewHexOrDecimal256(1),
			"gasToken":       common.HexToAddress("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC"),
			"refundReceiver": common.HexToAddress("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC"),
			"nonce":          math.NewHexOrDecimal256(1),
		},
	}

	fmt.Println(utils.JsonPretty(data))
	rsp, err := _cli_.EIP712Sign(context.Background(), "0xa1ec06ef11eb47f98bd919b5bf27c2107e44fd9f", big.NewInt(1337), data, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))
}

func TestCubist_PsbtSignMultisigTx(t *testing.T) {
	pks := []string{
		"tb1pcxvfy4ezehsha2h30dr9lpa8psg8axat6c0fssflwkpanlgwr7csym0cye",
		"tb1ps2zjfxsvalaqjyge9s4cz2c5l2mxf6augu7vez40m966qxfamqjsdjwk2l",
		//"tb1pzztghz3vlg7c07pd6fn48lchw9xsvux7gc2tluqaudg78hr97fmqews5sc",
	}

	inputPsbt := "70736274ff01005e020000000158f635405a3365d65f5a516c710e5c6c2efb3892a203de8a0f14e5e181049a5d0000000000ffffffff01084c010000000000225120b9e42233471b8d58e9cb98fda7ae088f16ff6fe71733c061ce73ea96ec613cd6000000000001012ba086010000000000225120acac41d12cb3a808e9fad77d531210298b4c5f4e3a1548ec3db2841ca740ab3d2215c150929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac06920be7c61b415dcf3117992b2c7941293db04f1723fea9a216d4427ee3d54d21d32ac20c81b336d1fdef81309eb4b6b9e073d9b70638cc76700cdb5276cd1078800ef0dba20f84a799e9873f9a8c19cd34007bc4c2173e48be882f3a50c949210903e314a4eba52a2c00000"

	for _, pk := range pks {
		rsp, err := _cli_.PsbtSign(context.Background(),
			pk,
			inputPsbt,
			nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(rsp)
		inputPsbt = rsp
	}

	signedPsbtBz, err := hex.DecodeString(inputPsbt)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := psbt.NewFromRawBytes(bytes.NewReader(signedPsbtBz), false)
	if err != nil {
		t.Fatal(err)
	}

	/*
	   {
	     const leafHash = tapleafHash({
	       output: leafScript,
	       version: LEAF_VERSION_TAPSCRIPT,
	     })
	     for (const input of psbt.data.inputs) {
	       if (!input.tapScriptSig) continue
	       const signedPubkeys = input.tapScriptSig.filter((ts) => ts.leafHash.equals(leafHash)).map((ts) => ts.pubkey)
	       for (const pubkey of leafPubkeys) {
	         if (signedPubkeys.some((sPub) => sPub.equals(pubkey))) continue
	         input.tapScriptSig.push({
	           // This can be reused for each dummy signature
	           leafHash,
	           // This is the pubkey that didn't sign
	           pubkey,
	           // This must be an empty Buffer.
	           signature: Buffer.from([]),
	         })
	       }
	     }
	   }
	*/
	if err = psbt.MaybeFinalizeAll(packet); err != nil {
		t.Fatal(err)
	}
	msgTx, err := psbt.Extract(packet)
	if err != nil {
		t.Fatal(err)
	}

	txid := msgTx.TxHash().String()
	var signedTxBuf bytes.Buffer
	if err := msgTx.Serialize(&signedTxBuf); err != nil {
		t.Fatal(err)
	}
	signedTx := hex.EncodeToString(signedTxBuf.Bytes())

	var multi_expporer_api = []string{"https://mempool.space:443/testnet/api/"}
	cli := bitcoinlib.NewSmartClient(multi_expporer_api)
	txid, err = cli.PostTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("txid:", txid)
}

func TestCubist_Eth1Sign(t *testing.T) {
	cli, err := ethclient.DialContext(context.Background(), "https://eth-sepolia.g.alchemy.com/v2/y8ILoqRkHr70iYwBIDLf0JMREp-5fHlI")
	if err != nil {
		t.Fatal(err)
	}

	number, err := cli.BlockNumber(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	block, err := cli.BlockByNumber(context.Background(), big.NewInt(int64(number)))
	if err != nil {
		t.Fatal(err)
	}
	baseFee := block.BaseFee()
	fmt.Println("baseFee", baseFee)
	tipFee, err := cli.SuggestGasTipCap(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	signer := "0x883f5d721c4653c37d64e4d7a301e90810cdab82"
	chainId := big.NewInt(11155111)

	nonce, err := cli.NonceAt(context.Background(), common.HexToAddress(signer), nil)
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := _cli_.Eth1Sign(context.Background(), signer,
		chainId,
		TransactionV2{
			ChainID:              hexutil.EncodeBig(chainId),
			From:                 "0x883f5d721c4653c37d64e4d7a301e90810cdab82",
			Gas:                  hexutil.EncodeBig(big.NewInt(21000)),
			MaxFeePerGas:         hexutil.EncodeBig(new(big.Int).Add(baseFee, tipFee)),
			MaxPriorityFeePerGas: hexutil.EncodeBig(tipFee),
			Nonce:                hexutil.EncodeBig(big.NewInt(int64(nonce))),
			To:                   "0x800F9c6fbcD0F7C78dA3AA58C83cCc6356C5Cf8f",
			Type:                 "0x02",
			Value:                hexutil.EncodeBig(big.NewInt(100000000000000)),
			Data:                 "",
			AccessList:           nil,
		},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.JsonPretty(rsp))

	var tx = types.Transaction{}
	if err := tx.UnmarshalBinary(hexutil.MustDecode(rsp)); err != nil {
		t.Fatal(err)
	}
	t.Log(tx.Hash().Hex())
	if err := cli.SendTransaction(context.TODO(), &tx); err != nil {
		t.Fatal(err)
	}
}

func TestCubist_SolanaSign(t *testing.T) {
	messageB64 := "gAEAAQMdt2FcaOPvKa5uIX8X7dxsjnDfbpK5azBJUAbT8yBoanlybaUtmdYLB+rXOy9vC/YIPMhcd6lONNaR14+Lyv7JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC6Movkc30WVHA20gz0b8/waop8/0Nrue423h8mBB8lMAECAgABDAIAAABkAAAAAAAAAAA="
	signer := "3111LEM8qzzeDW2maez6KuXZ49X9QTxhXBe4MKJAzXW9"
	rsp, err := _cli_.SolanaSign(context.Background(), signer,
		messageB64,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rsp)
}
