package dynamic_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testLogger struct {
	t    *testing.T
	name string
}

func newTestLogger(t *testing.T, name string) *testLogger {
	return &testLogger{t: t, name: name}
}

func (tl *testLogger) SetName(name string) {
	tl.name = name
}
func (tl *testLogger) LogDebug(format string, a ...interface{}) {
	tl.t.Logf(format, a...)
}
func (tl *testLogger) LogInfo(format string, a ...interface{}) {
	tl.t.Logf(format, a...)
}
func (tl *testLogger) LogError(err error, format string, a ...interface{}) {
	tl.t.Logf("error: %s."+format, err, a)
}

func (tl *testLogger) ErrUserAuth(c context.Context, err error) error {
	tl.t.Logf("error: %s", err.Error())
	return nil
}
func (tl *testLogger) InfoGeneric(c context.Context, s string) error {
	tl.t.Logf("info: %s", s)
	return nil
}
func (tl *testLogger) ErrGeneric(c context.Context, err error) error {
	tl.t.Logf("error: %s", err.Error())
	return nil
}

// generate is copied from gesundheitscloud/jwt-rotator
func generate(t *testing.T, keyLen int) (public, private string) {
	key, err := rsa.GenerateKey(rand.Reader, keyLen)
	if err != nil {
		t.Fatal("generating key-pair failed")
	}
	keyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal("error when running MarshalPKIXPublicKey: %w", err)
	}
	pubkey_pem := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: keyBytes}))
	privatekey_pem := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	return pubkey_pem, privatekey_pem
}

// keyEntries is test helper (hence t in params) to generate a pair of valid key entries
func keyEntries(t *testing.T, name string, enabled bool) (privEntry, pubEntry string) {
	pub, priv := generate(t, 2048)
	privEntry = `
- name: "` + name + `"
  comment: "generated with: rsa.GenerateKey()"
  enabled: ` + fmt.Sprintf("%t", enabled) + `
  created_at: "unknown"
  author: "Jenkins"
  key: |
    ` + strings.ReplaceAll(priv, "\n", "\n    ")
	pubEntry = `
- name: "` + name + `"
  comment: "generated with: rsa.GenerateKey() and MarshalPKIXPublicKey"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    ` + strings.ReplaceAll(pub, "\n", "\n    ")
	return privEntry, pubEntry
}

func publicKeyYaml(entries ...string) string {
	var yamlExample = `
JWTPublicKey:
`
	if len(entries) == 0 {
		return `JWTPublicKey: []`
	}
	for _, e := range entries {
		yamlExample = yamlExample + "\n" + e
	}
	return yamlExample
}

func privateKeyYaml(entries ...string) string {
	var yamlExample = `
JWTPrivateKey:
`
	if len(entries) == 0 {
		return `JWTPrivateKey: []`
	}
	for _, e := range entries {
		yamlExample = yamlExample + "\n" + e
	}
	return yamlExample
}

func configYaml(pub, priv string) string {
	return pub + `
` + priv
}

func generateAndWriteConfig(t *testing.T, numKeys int, namePrefix string, dir string) string {
	pubs := make([]string, numKeys)
	privs := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		enabled := i == 0
		priv, pub := keyEntries(t, fmt.Sprintf("%s%d", namePrefix, i), enabled)
		pubs = append(pubs, pub)
		privs = append(privs, priv)
	}
	var yaml = []byte(configYaml(publicKeyYaml(pubs...), privateKeyYaml(privs...)))

	var tmpDir string
	// create a tmp dir if it does not exist yet
	if dir == "" {
		d, err := os.MkdirTemp("", "unittest")
		if err != nil {
			t.Fatal(err)
		}
		tmpDir = d
	} else {
		tmpDir = dir
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), yaml, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir, "config.yaml"))
	return tmpDir
}

// testRequest is copied from https://github.com/go-chi/chi/blob/d32a83448b5f43e42bc96487c6b0b3667a92a2e4/middleware/middleware_test.go#L83
func testRequest(t *testing.T, ts *httptest.Server, method, path string, token string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
