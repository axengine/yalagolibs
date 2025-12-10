package cubist

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/axengine/utils"
	"github.com/go-resty/resty/v2"
)

const Tips = `make sure signer-session.json is in the dir`

type Cubist struct {
	debug    bool
	dir      string
	signerMu sync.Mutex
	orgId    string
}

func New(debug bool, dir string) *Cubist {
	return &Cubist{
		debug: debug,
		dir:   dir,
	}
}

func (c *Cubist) Init(ctx context.Context) error {
	sess, err := c.loadSignerSession(ctx)
	if sess != nil {
		c.orgId = sess.OrgID
	}
	return err
}

func (c *Cubist) Refresh(ctx context.Context, interval time.Duration) {
	tk := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("cubist refresh session coroutine exit")
			return
		case <-tk.C:
			if err := c.Init(ctx); err != nil {
				log.Println("cubist init error", err)
				continue
			}
			log.Println("cubist refresh session success")
		}
	}
}

func (c *Cubist) loadSignerSession(ctx context.Context) (*Session, error) {
	c.signerMu.Lock()
	defer c.signerMu.Unlock()
	session, err := loadSignerSession(c.dir)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() > session.SessionInfo.AuthTokenExp {
		if err := c.refreshToken(ctx, session); err != nil {
			return nil, err
		}
		if err := updateSignerSession(session, c.dir); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (c *Cubist) refreshToken(ctx context.Context, session *Session) error {
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)

	ro := make(map[string]interface{})
	ro["epoch_num"] = session.SessionInfo.Epoch
	ro["epoch_token"] = session.SessionInfo.EpochToken
	ro["other_token"] = session.SessionInfo.RefreshToken

	uri := fmt.Sprintf("/v1/org/%s/token/refresh", session.OrgID)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetHeader("Content-Type", "application/json").SetBody(ro).Patch(uri)
	if err != nil {
		return err
	}
	if rsp.StatusCode() != 200 {
		return fmt.Errorf("refresh token error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data Session
	if err = json.Unmarshal(rsp.Body(), &data); err != nil {
		return err
	}

	session.Token = data.Token
	session.RefreshToken = data.RefreshToken
	session.Expiration = data.Expiration
	session.SessionInfo = data.SessionInfo
	return nil
}

func (c *Cubist) getMfa(ctx context.Context, mfaId string) (*MfaRequest, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return nil, err
	}
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	uri := fmt.Sprintf("/v0/org/%s/mfa/%s", session.OrgID, mfaId)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("get mfa error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = MfaRequest{}
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Cubist) signAny(ctx context.Context, uri string, ro any, headers map[string]string) (*SignResponse, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return nil, err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)
	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	rsp, err := r.SetBody(ro).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() == 202 { //MFA required
		var data = MFARequiredResponse{}
		if err := json.Unmarshal(rsp.Body(), &data); err != nil {
			return nil, err
		}
		tk := time.NewTimer(time.Second * 5)
		tm := time.NewTimer(time.Minute * 5) // max 5mins wait to approved
		for {
			select {
			case <-ctx.Done():
				return nil, errors.New("timeout to cancel")
			case <-tk.C:
				mfaRequest, err := c.getMfa(ctx, data.Accepted.MfaRequired.ID)
				if err != nil {
					return nil, err
				}
				log.Println("got mfa:", utils.JsonPretty(mfaRequest))
				if mfaRequest.Receipt != nil {
					receipt := make(map[string]interface{})
					receipt["id"] = data.Accepted.MfaRequired.ID
					receipt["confirmation"] = mfaRequest.Receipt.Confirmation

					var receipts []map[string]interface{}
					receipts = append(receipts, receipt)
					bz, _ := json.Marshal(receipts)
					headers = make(map[string]string)
					headers["x-cubist-mfa-org-id"] = data.Accepted.MfaRequired.OrgID
					headers["x-cubist-mfa-receipts"] = encodeToBase64Url(bz)
					return c.signAny(ctx, uri, ro, headers)
				}
				tk.Reset(time.Second * 5)
			case <-tm.C:
				return nil, errors.New("timeout to wait approved")
			}
		}
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var rlt SignResponse
	if err := json.Unmarshal(rsp.Body(), &rlt); err != nil {
		return nil, err
	}
	return &rlt, nil
}

func (c *Cubist) SegwitSignV0(ctx context.Context, pubkey string, ro *SegwitSignRo) (string, error) {
	uri := fmt.Sprintf("/v0/org/%s/btc/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Signature, nil
}

func (c *Cubist) SegwitSign(ctx context.Context, pubkey string, ro *SegwitSignRo, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	uri := fmt.Sprintf("/v0/org/%s/btc/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(ro).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("segwit sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = struct {
		Signature string `json:"signature"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return "", err
	}
	return data.Signature, nil
}

func (c *Cubist) PsbtSignV0(ctx context.Context, pubkey string, psbt string) (string, error) {
	uri := fmt.Sprintf("/v0/org/%s/btc/psbt/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	ro := map[string]interface{}{
		"psbt":             psbt,
		"sign_all_scripts": true,
	}
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Psbt, nil
}

// PsbtSign signs the PSBT and appends the specified headers
func (c *Cubist) PsbtSign(ctx context.Context, pubkey string, psbt string, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	uri := fmt.Sprintf("/v0/org/%s/btc/psbt/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(map[string]interface{}{
		"psbt":             psbt,
		"sign_all_scripts": true,
	}).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() == 202 {

	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("psbt sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = struct {
		Psbt string `json:"psbt"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return "", err
	}
	return data.Psbt, nil
}

func (c *Cubist) EIP191SignV0(ctx context.Context, pubkey string, data string) (string, error) {
	uri := fmt.Sprintf("/v0/org/%s/evm/eip191/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	ro := map[string]interface{}{
		"data": data,
	}
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Signature, nil
}

func (c *Cubist) EIP191Sign(ctx context.Context, pubkey string, data string, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	uri := fmt.Sprintf("/v0/org/%s/evm/eip191/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(map[string]interface{}{
		"data": data,
	}).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("eip191 sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var signature = struct {
		Signature string `json:"signature"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &signature); err != nil {
		return "", err
	}
	return signature.Signature, nil
}

func (c *Cubist) EIP712SignV0(ctx context.Context, pubkey string, chainId *big.Int, typedData *TypedData) (string, error) {
	uri := fmt.Sprintf("/v0/org/%s/evm/eip712/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	ro := struct {
		ChainId   int64      `json:"chain_id"`
		TypedData *TypedData `json:"typed_data"`
	}{
		chainId.Int64(), typedData,
	}
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Signature, nil
}

func (c *Cubist) EIP712Sign(ctx context.Context, pubkey string, chainId *big.Int, typedData *TypedData, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	ro := struct {
		ChainId   int64      `json:"chain_id"`
		TypedData *TypedData `json:"typed_data"`
	}{
		chainId.Int64(), typedData,
	}
	//bz, _ := json.MarshalIndent(&ro, "", "  ")
	//fmt.Println(string(bz))

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	uri := fmt.Sprintf("/v0/org/%s/evm/eip712/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(ro).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("eip712 sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var signature = struct {
		Signature string `json:"signature"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &signature); err != nil {
		return "", err
	}
	return signature.Signature, nil
}

func (c *Cubist) Eth1SignV0(ctx context.Context, pubkey string, chainId *big.Int, txData interface{}) (string, error) {
	uri := fmt.Sprintf("/v1/org/%s/eth1/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	ro := struct {
		ChainId int64       `json:"chain_id"`
		Tx      interface{} `json:"tx"`
	}{
		chainId.Int64(), txData,
	}
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Signature, nil
}

func (c *Cubist) Eth1Sign(ctx context.Context, pubkey string, chainId *big.Int, txData interface{}, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)
	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}

	ro := struct {
		ChainId int64       `json:"chain_id"`
		Tx      interface{} `json:"tx"`
	}{
		chainId.Int64(), txData,
	}
	// bz, _ := json.MarshalIndent(&ro, "", "  ")
	// fmt.Println(string(bz))

	uri := fmt.Sprintf("/v1/org/%s/eth1/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(ro).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("eth1 sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var retult = struct {
		RLPSignedTx string `json:"rlp_signed_tx"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &retult); err != nil {
		return "", err
	}
	return retult.RLPSignedTx, nil
}

func (c *Cubist) SolanaSignV0(ctx context.Context, pubkey string, base64 string) (string, error) {
	uri := fmt.Sprintf("/v1/org/%s/solana/sign/%s", c.orgId, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	ro := map[string]interface{}{
		"message_base64": base64,
	}
	rlt, err := c.signAny(ctx, uri, ro, nil)
	if err != nil {
		return "", err
	}
	return rlt.Signature, nil
}

func (c *Cubist) SolanaSign(ctx context.Context, pubkey string, base64 string, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession(ctx)
	if err != nil {
		return "", err
	}

	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	uri := fmt.Sprintf("/v0/org/%s/solana/sign/%s", session.OrgID, pubkey)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.SetBody(map[string]interface{}{
		"message_base64": base64,
	}).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode() != 200 {
		return "", fmt.Errorf("solana sign error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = struct {
		Signature string `json:"signature"`
	}{}
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return "", err
	}
	return data.Signature, nil
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
