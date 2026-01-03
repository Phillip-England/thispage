package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type storedCredentials struct {
	Version    int    `json:"version"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

type plainCredentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	SessionToken string `json:"session_token"`
}

func projectSeedPath(projectPath string) string {
	return filepath.Join(projectPath, ".thispage", "seed")
}

func ProjectSeed(projectPath string) (string, error) {
	seedPath := projectSeedPath(projectPath)
	data, err := os.ReadFile(seedPath)
	if err != nil {
		return "", fmt.Errorf("project seed file is missing: %w", err)
	}

	seed := strings.TrimSpace(string(data))
	if seed == "" {
		return "", fmt.Errorf("project seed file is empty: %s", seedPath)
	}

	return seed, nil
}

func EnsureProjectSeed(projectPath string) (string, error) {
	if seed, err := ProjectSeed(projectPath); err == nil {
		return seed, nil
	}

	seedPath := projectSeedPath(projectPath)

	seedBytes := make([]byte, 32)
	if _, err := rand.Read(seedBytes); err != nil {
		return "", fmt.Errorf("failed to generate seed: %w", err)
	}
	seed := base64.RawStdEncoding.EncodeToString(seedBytes)

	if err := os.MkdirAll(filepath.Dir(seedPath), 0700); err != nil {
		return "", fmt.Errorf("failed to create seed directory: %w", err)
	}
	if err := os.WriteFile(seedPath, []byte(seed), 0600); err != nil {
		return "", fmt.Errorf("failed to write seed file: %w", err)
	}

	return seed, nil
}

func Save(projectPath, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	seed, err := EnsureProjectSeed(projectPath)
	if err != nil {
		return err
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return err
	}

	return saveWithToken(projectPath, username, password, token, seed)
}

func saveWithToken(projectPath, username, password, token, seed string) error {
	payload := plainCredentials{
		Username:     username,
		Password:     password,
		SessionToken: token,
	}
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	key := sha256.Sum256([]byte(seed))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	encoded := storedCredentials{
		Version:    1,
		Nonce:      base64.RawStdEncoding.EncodeToString(nonce),
		Ciphertext: base64.RawStdEncoding.EncodeToString(ciphertext),
	}

	data, err := json.MarshalIndent(encoded, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize encrypted credentials: %w", err)
	}

	credPath := credentialsPath(projectPath)
	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

func Load(projectPath string) (string, string, error) {
	username, password, _, err := LoadWithToken(projectPath)
	return username, password, err
}

func LoadWithToken(projectPath string) (string, string, string, error) {
	seed, err := ProjectSeed(projectPath)
	if err != nil {
		return "", "", "", err
	}

	credPath := credentialsPath(projectPath)
	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read credentials: %w", err)
	}

	var stored storedCredentials
	if err := json.Unmarshal(data, &stored); err != nil {
		return "", "", "", fmt.Errorf("failed to parse credentials: %w", err)
	}
	if stored.Ciphertext == "" || stored.Nonce == "" {
		return "", "", "", fmt.Errorf("credentials file is missing required fields")
	}

	nonce, err := base64.RawStdEncoding.DecodeString(stored.Nonce)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decode nonce: %w", err)
	}
	ciphertext, err := base64.RawStdEncoding.DecodeString(stored.Ciphertext)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	key := sha256.Sum256([]byte(seed))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var decoded plainCredentials
	if err := json.Unmarshal(plaintext, &decoded); err != nil {
		return "", "", "", fmt.Errorf("failed to parse decrypted credentials: %w", err)
	}

	if decoded.SessionToken == "" {
		token, err := GenerateSessionToken()
		if err != nil {
			return "", "", "", err
		}
		if err := saveWithToken(projectPath, decoded.Username, decoded.Password, token, seed); err != nil {
			return "", "", "", err
		}
		decoded.SessionToken = token
	}

	return decoded.Username, decoded.Password, decoded.SessionToken, nil
}

func credentialsPath(projectPath string) string {
	return filepath.Join(projectPath, ".thispage", "credentials.json")
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(bytes), nil
}
