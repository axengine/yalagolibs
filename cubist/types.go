package cubist

type MFARequiredResponse struct {
	Message  string `json:"message"`
	Accepted struct {
		MfaRequired struct {
			ID    string   `json:"id"`
			IDs   []string `json:"ids"`
			OrgID string   `json:"org_id"`
		} `json:"MfaRequired"`
	} `json:"accepted"`
	RequestID string `json:"request_id"`
	ErrorCode string `json:"error_code"`
}

type MfaRequest struct {
	CreatedAt  int64    `json:"created_at"`
	CreatedBy  string   `json:"created_by"`
	ExpiresAt  int64    `json:"expires_at"`
	ID         string   `json:"id"`
	Provenance string   `json:"provenance"`
	Receipt    *Receipt `json:"receipt"`
	RelatedIDs []string `json:"related_ids"`
	Request    Request  `json:"request"`
	Status     Status   `json:"status"`
}

type SignResponse struct {
	Signature string `json:"signature,omitempty"` // for psbt sign
	Psbt      string `json:"psbt,omitempty"`      // for others sign
}

type Status struct {
	AllowedApprovers []string `json:"allowed_approvers"`
	AllowedMfaTypes  []string `json:"allowed_mfa_types"`
	ApprovedBy       any      `json:"approved_by"`
	Count            int      `json:"count"`
	NumAuthFactors   int      `json:"num_auth_factors"`
	RequestComparer  any      `json:"request_comparer"`
}

type Request struct {
	Body   any    `json:"body"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Receipt struct {
	Confirmation  string `json:"confirmation"`
	FinalApprover string `json:"final_approver"`
	Timestamp     int64  `json:"timestamp"`
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
