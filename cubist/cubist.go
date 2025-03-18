package cubist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const Tips = `
1、check signer-session.json exist?
2、if not,please login with command 'cs'
3、check signer-session.json expired，if expired,delete it
4、check the key has sign scope sign:bitcoin:psbt or sign:evm:eip191
`

type Cubist struct {
	debug bool
}

func New(debug bool) *Cubist {
	return &Cubist{
		debug: debug,
	}
}

func (c *Cubist) Init() error {
	_, err := c.loadSignerSession()
	return err
}

func (c *Cubist) Refresh(ctx context.Context, wg *sync.WaitGroup, interval time.Duration) {
	defer wg.Done()
	tk := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("cubist refresh session coroutine exit")
			return
		case <-tk.C:
			if err := c.Init(); err != nil {
				log.Println("cubist init error", zap.Error(err))
				continue
			}
			log.Println("cubist refresh session success")
		}
	}
}

func (c *Cubist) loadManagementSession() (*Session, error) {
	session, err := loadManagementSession()
	if err != nil {
		return nil, err
	}
	if time.Now().Unix() > session.SessionInfo.AuthTokenExp {
		if err := c.refreshToken(session); err != nil {
			return nil, err
		}
		if err := updateManagementSession(session); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (c *Cubist) loadSignerSession() (*Session, error) {
	session, err := loadSignerSession()
	if err != nil {
		session, err = c.createSession(context.Background())
		if err != nil {
			return nil, err
		}
		if err := updateSignerSession(session); err != nil {
			return nil, err
		}
	}

	if time.Now().Unix() > session.SessionInfo.AuthTokenExp {
		if err := c.refreshToken(session); err != nil {
			return nil, err
		}
		if err := updateSignerSession(session); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (c *Cubist) refreshToken(session *Session) error {
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetHeader("Authorization", session.Token)

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

func (c *Cubist) Me() (interface{}, error) {
	session, err := c.loadManagementSession()
	if err != nil {
		return nil, err
	}
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetHeader("Authorization", session.Token)
	uri := fmt.Sprintf("/v0/org/%s/user/me", session.OrgID)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("me error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var me = make(map[string]interface{})
	if err := json.Unmarshal(rsp.Body(), &me); err != nil {
		return nil, err
	}
	return me, nil
}

func (c *Cubist) Org() (interface{}, error) {
	session, err := c.loadManagementSession()
	if err != nil {
		return nil, err
	}
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetHeader("Authorization", session.Token)
	uri := fmt.Sprintf("/v0/org/%s", session.OrgID)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("org error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = make(map[string]interface{})
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Cubist) Keys() (interface{}, error) {
	session, err := c.loadManagementSession()
	if err != nil {
		return nil, err
	}
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetHeader("Authorization", session.Token)
	uri := fmt.Sprintf("/v0/org/%s/keys", session.OrgID)
	uri = strings.Replace(uri, "#", "%23", -1)
	rsp, err := r.Get(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("keys error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data = make(map[string]interface{})
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return nil, err
	}
	return data, nil
}

type Segwit struct {
	InputIndex  int    `json:"input_index"`
	ScriptCode  string `json:"script_code"`
	SighashType string `json:"sighash_type"`
	Value       int64  `json:"value"`
}
type SigKind struct {
	Segwit Segwit `json:"Segwit"`
}

type Input struct {
	PreviousOutput string   `json:"previous_output"`
	ScriptSig      string   `json:"script_sig"`
	Sequence       uint32   `json:"sequence"`
	Witness        []string `json:"witness"`
}

type Output struct {
	ScriptPubkey string `json:"script_pubkey"`
	Value        int64  `json:"value"`
}

type BitcoinTx struct {
	Version  int32    `json:"version"`
	Locktime uint32   `json:"lock_time"`
	Input    []Input  `json:"input"`
	Output   []Output `json:"output"`
}
type SegwitSignRo struct {
	SigKind SigKind   `json:"sig_kind"`
	Tx      BitcoinTx `json:"tx"`
}

func (c *Cubist) SegwitSign(ctx context.Context, pubkey string, ro *SegwitSignRo, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession()
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

// PsbtSign signs the PSBT and appends the specified headers
func (c *Cubist) PsbtSign(ctx context.Context, pubkey string, psbt string, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession()
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

func (c *Cubist) EIP191Sign(ctx context.Context, pubkey string, data string, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession()
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

func (c *Cubist) EIP712Sign(ctx context.Context, pubkey string, chainId *big.Int, typedData *TypedData, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession()
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

func (c *Cubist) Eth1Sign(ctx context.Context, pubkey string, chainId *big.Int, txData interface{}, headers map[string]string) (string, error) {
	session, err := c.loadSignerSession()
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
	bz, _ := json.MarshalIndent(&ro, "", "  ")
	fmt.Println(string(bz))

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

func (c *Cubist) createSession(ctx context.Context) (*Session, error) {
	session, err := c.loadManagementSession()
	if err != nil {
		return nil, err
	}
	cli := resty.New().SetBaseURL(session.Env.DevCubeSignerStack.SignerApiRoot)

	r := cli.R().SetContext(ctx).SetHeader("Authorization", session.Token)
	uri := fmt.Sprintf("/v0/org/%s/session", session.OrgID)
	uri = strings.Replace(uri, "#", "%23", -1)

	ro := make(map[string]interface{})
	ro["purpose"] = "auto sign"
	ro["scopes"] = []string{"manage:key:get", "sign:btc:segwit", "sign:btc:psbt:*", "sign:evm:eip712", "sign:evm:tx"}
	if c.debug {
		ro["auth_lifetime"] = 3000        // 5mins
		ro["refresh_lifetime"] = 86400    // 1day
		ro["session_lifetime"] = 31536000 // 1year
		ro["grace_lifetime"] = 30         // 30s
	} else {
		ro["auth_lifetime"] = 300       // 5mins
		ro["refresh_lifetime"] = 86400  // 1day
		ro["session_lifetime"] = 604800 // 7days
		ro["grace_lifetime"] = 30       // 30s
	}
	rsp, err := r.SetBody(ro).SetHeader("Content-Type", "application/json").Post(uri)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, fmt.Errorf("create session error,status:%s message:%s", rsp.Status(), rsp.String())
	}
	var data Session
	if err := json.Unmarshal(rsp.Body(), &data); err != nil {
		return nil, err
	}
	data.OrgID = session.OrgID
	data.RoleID = session.RoleID
	data.Purpose = session.Purpose
	data.Env = session.Env

	return &data, nil
}
