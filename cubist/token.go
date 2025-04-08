package cubist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Env represents the environment configuration.
type Env struct {
	DevCubeSignerStack struct {
		ClientId                 string      `json:"ClientId"`
		GoogleDeviceClientId     string      `json:"GoogleDeviceClientId"`
		GoogleDeviceClientSecret string      `json:"GoogleDeviceClientSecret"`
		Region                   string      `json:"Region"`
		UserPoolId               string      `json:"UserPoolId"`
		SignerApiRoot            string      `json:"SignerApiRoot"`
		DefaultCredentialRpId    string      `json:"DefaultCredentialRpId"`
		EncExportS3BucketName    interface{} `json:"EncExportS3BucketName"`   // Optional field, hence the pointer
		DeletedKeysS3BucketName  interface{} `json:"DeletedKeysS3BucketName"` // Optional field, hence the pointer
	} `json:"Dev-CubeSignerStack"`
}

// SessionInfo holds the session-related information.
type SessionInfo struct {
	AuthToken       string `json:"auth_token"`
	AuthTokenExp    int64  `json:"auth_token_exp"`
	Epoch           int    `json:"epoch"`
	EpochToken      string `json:"epoch_token"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp int64  `json:"refresh_token_exp"`
	SessionID       string `json:"session_id"`
}

// Session holds the entire configuration structure.
type Session struct {
	OrgID        string      `json:"org_id"`
	RoleID       string      `json:"role_id"`
	Expiration   int64       `json:"expiration"`
	Purpose      string      `json:"purpose"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	Env          Env         `json:"env"`
	SessionInfo  SessionInfo `json:"session_info"`
}

func configDir() string {
	var configDir string
	// Check the operating system
	if runtime.GOOS == "darwin" {
		// For macOS
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	} else {
		// For other operating systems (Linux, Windows, etc.)
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	// Append the application-specific directory
	return filepath.Join(configDir, "cubesigner")
}

var _CONFIG_DIR = configDir()
var _MANAGEMENT_SESSION_PATH = filepath.Join(_CONFIG_DIR, "management-session.json")
var _SIGNER_SESSION_PATH = filepath.Join(_CONFIG_DIR, "signer-session.json")

func loadManagementSession(dir string) (*Session, error) {
	if dir != "" {
		_MANAGEMENT_SESSION_PATH = filepath.Join(dir, "management-session.json")
	}
	bz, err := os.ReadFile(_MANAGEMENT_SESSION_PATH)
	if err != nil {
		return nil, err
	}
	var session Session
	if err = json.Unmarshal(bz, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func updateManagementSession(session *Session, dir string) error {
	if dir != "" {
		_MANAGEMENT_SESSION_PATH = filepath.Join(dir, "management-session.json")
	}
	bz, _ := json.Marshal(&session)
	return os.WriteFile(_MANAGEMENT_SESSION_PATH, bz, 0755)
}

func loadSignerSession(dir string) (*Session, error) {
	if dir != "" {
		_SIGNER_SESSION_PATH = filepath.Join(dir, "signer-session.json")
	}
	bz, err := os.ReadFile(_SIGNER_SESSION_PATH)
	if err != nil {
		return nil, err
	}
	var session Session
	if err = json.Unmarshal(bz, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func updateSignerSession(session *Session, dir string) error {
	if dir != "" {
		_SIGNER_SESSION_PATH = filepath.Join(dir, "signer-session.json")
	}
	bz, _ := json.Marshal(&session)
	return os.WriteFile(_SIGNER_SESSION_PATH, bz, 0755)
}
