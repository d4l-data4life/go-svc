package dynamic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestViperConfigCopying ensures that ViperConfing object works correctly when copied in various ways (as long as 'go vet' allows)
func TestViperConfigCopying(t *testing.T) {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample1 = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "public1") + `
` + pubKeyEntry(t, "public2") + `
JWTPrivateKey:
` + privKeyEntry(t, "private1", true) + `
` + privKeyEntry(t, "private2", false) + `
`)
	var yamlExample2 = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "publicA") + `
JWTPrivateKey:
` + privKeyEntry(t, "privateA", true) + `
`)
	// write config 1 to temp file
	tmpDir, err := os.MkdirTemp("", "unittest")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("created dir %s", tmpDir)
	defer func() {
		t.Logf("deleting dir %s", tmpDir)
		os.RemoveAll(tmpDir) // clean up
	}()
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "config.yaml"), yamlExample1, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir, "config.yaml"))

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigFileName("config"),
		WithConfigFilePaths(tmpDir),
		WithAutoBootstrap(true),
		WithWatchChanges(true),
	)
	if vc.Error != nil {
		log.Fatal(err, "unable to bootstrap VC")
	}

	arr, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "private1", pkey.Name)

	// now get an interface
	type vcInterface interface {
		JWTPublicKeys() ([]JWTPublicKey, error)
		JWTPrivateKey() (JWTPrivateKey, error)
	}

	// interfacerFun accepts VC in form of an interface
	interfacerFun := func(d vcInterface, t *testing.T, numPub int, privName, testName string) {
		arr, err := d.JWTPublicKeys()
		assert.NoError(t, err, testName)
		assert.Equal(t, numPub, len(arr), testName)

		pkey, err := d.JWTPrivateKey()
		assert.NoError(t, err, testName)
		assert.Equal(t, privName, pkey.Name, testName)
	}
	// we also tested this for copying of vc, but 'go vet' disallows such code -
	// call of purerFun copies lock value: (...)dynamic.ViperConfig contains sync.Once contains sync.Mutex

	// pointerFun accepts VC in its natural form as pointer
	pointerFun := func(vc *ViperConfig, t *testing.T, numPub int, privName, testName string) {
		arr, err := vc.JWTPublicKeys()
		assert.NoError(t, err, testName)
		assert.Equal(t, numPub, len(arr), testName)

		pkey, err := vc.JWTPrivateKey()
		assert.NoError(t, err, testName)
		assert.Equal(t, privName, pkey.Name, testName)
	}

	interfacerFun(vc, t, 2, "private1", "phase1-interfacer")
	pointerFun(vc, t, 2, "private1", "phase1-pointer")

	// now simulate hot-reload of the config
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "config.yaml"), yamlExample2, 0666); err != nil {
		t.Fatal(err)
	}
	// wait a bit for the filesystem to catch the changes in the config files and notify Viper
	<-time.After(50 * time.Millisecond)

	// all forms should catch the new changes
	interfacerFun(vc, t, 1, "privateA", "phase2-interfacer")
	pointerFun(vc, t, 1, "privateA", "phase2-pointer")
}

// TestViperConfigHotReloadAfterMerge ensures that ViperConfing (VC) object works (in)correctly after merging with another VC object
func TestViperConfigHotReloadAfterMerge(t *testing.T) {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample1 = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "public1") + `
` + pubKeyEntry(t, "public2"))

	var yamlExample2 = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "private1", true) + `
` + privKeyEntry(t, "private2", false) + `
`)
	// prepare two separate temporary dirs
	tmpDir1, err := os.MkdirTemp("", "unittest1")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("created dir %s", tmpDir1)
	defer func() {
		t.Logf("deleting dir %s", tmpDir1)
		os.RemoveAll(tmpDir1) // clean up
	}()
	tmpDir2, err := os.MkdirTemp("", "unittest2")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("created dir %s", tmpDir2)
	defer func() {
		t.Logf("deleting dir %s", tmpDir2)
		os.RemoveAll(tmpDir2) // clean up
	}()

	// Write to config.yaml files to separate dirs
	if err := ioutil.WriteFile(filepath.Join(tmpDir1, "config.yaml"), yamlExample1, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir1, "config.yaml"))

	if err := ioutil.WriteFile(filepath.Join(tmpDir2, "config.yaml"), yamlExample2, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir2, "config.yaml"))

	// build 2 separate vc objects
	vc1 := NewViperConfig("test1", WithConfigFormat("yaml"),
		WithConfigFileName("config"),
		WithConfigFilePaths(tmpDir1),
		WithAutoBootstrap(true),
		WithWatchChanges(true),
	)
	if vc1.Error != nil {
		log.Fatal(err, "unable to bootstrap VC1")
	}
	vc2 := NewViperConfig("test2", WithConfigFormat("yaml"),
		WithConfigFileName("config"),
		WithConfigFilePaths(tmpDir2),
		WithAutoBootstrap(true),
		WithWatchChanges(true),
	)
	if vc2.Error != nil {
		log.Fatal(err, "unable to bootstrap VC2")
	}
	// sanity checks
	arr, err := vc1.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))

	pkey, err := vc2.JWTPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "private1", pkey.Name)

	// now merge the two VC into 1
	err = vc1.MergeAndDisableHotReload(vc2)
	assert.NoError(t, err)
	// vc1 consits now of vc1+vc2
	assert.Equal(t, vc1.name, "merged-test1+test2")
	// now, the bound between merged vc1 and config file 2 is lost
	// updating config.yaml in tmpDir2 will NOT lead to correct hot-reloading of merged config, however, the config from tmpDir1 should be updated

	// now simulate hot-reload of the config 1 and 2
	var yamlExample1B = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "publicA") + `
` + pubKeyEntry(t, "publicB") + `
` + pubKeyEntry(t, "publicC"))

	if err := ioutil.WriteFile(filepath.Join(tmpDir1, "config.yaml"), yamlExample1B, 0666); err != nil {
		t.Fatal(err)
	}
	var yamlExample2B = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "privateA", true) + `
` + privKeyEntry(t, "privateB", false) + `
`)

	if err := ioutil.WriteFile(filepath.Join(tmpDir2, "config.yaml"), yamlExample2B, 0666); err != nil {
		t.Fatal(err)
	}
	// wait a bit for the filesystem to catch the changes in the config files and notify Viper
	<-time.After(50 * time.Millisecond)

	arr, err = vc1.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(arr))
	assert.Equal(t, "publicA", arr[0].Name)
	assert.Equal(t, "publicB", arr[1].Name)
	assert.Equal(t, "publicC", arr[2].Name)

	pkey, err = vc1.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotEqual(t, "privateA", pkey.Name) // We all would like that this would work this way ;(
	assert.Equal(t, "private1", pkey.Name)    // instead, the hot-reload for the vc2 is ignored, and we see the old value here

}

// privKeyEntry is test helper (hence t in params) to generate a valid priv Key entry
func privKeyEntry(t *testing.T, name string, enabled bool) string {
	return `
- name: "` + name + `"
  comment: "generated with: openssl genrsa 2048 -out private.pem"
  enabled: ` + fmt.Sprintf("%t", enabled) + `
  created_at: "unknown"
  author: "Jenkins"
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEArUG/lcVDen29RnHK+f1E5UzoXAAMTT5Xatdts7o4+jNwXl7L
    uYSAnmrl50XjyM5fLCog6G+qLz0L6U07EbXB0B/paHuYLlAG9rIYIaAceZYrhMRe
    USx3DL2yIpawa1YR9QYgyHTY2/3sXF+vd/T7JNqBxI/v0vZkhaFCugrWlvAICC1Y
    jQXrjZqRRPl0OsUwZ2kmRvlqvYcVSLEif+uKeNMyplThb9CEQZdjjLMSskYzcmGS
    fPc10ciEDhYR4O5M8vOO5DLeLwj6dw/PTrrslAxrdQqiP4/xyx89ZfFMsxBIBw5J
    eZ0VnQ46Chr5Dy34A/FacA3Sqb0ZEFkmwCTBjwIDAQABAoIBAQCMpb0zhinbPEv0
    7deKzVGqm55dYSSbaCpq72t85YXvhuaHlYjol2oaMElmT9Q0ZWPZZHHGfy+2nWYY
    BLwZCmXF4MIIMZ0+q3Sbu8PfOC0lfwThCNBQMTqLu0rqzU12NS7qrAjc8g5BuIay
    DnNRfCyMpF2IBhj4N1EvMdQLV1UQvYChvuok/oe9xxXPlhb9HCrHhs0WXamhuYmj
    2ZkAPtZ/zM7tzeiHfczx5sUh2BqtkiWDcpezkDhEQhn7C6b3C/2UGfQ8Q1CM3ske
    3D7uLSctvbWH3JNYm0QzRQUgXKYK9xfPsFVv7nxNZyOMtHIrary2Po6PfaGxkGvX
    sdRusDjBAoGBANdzbLNInge/wQKYeUJ7CoOcBWKIpa3kMIAy22wkSAFzE70gCHEn
    7/ppdUGmvHnuzULGQOtXkoHJ3S0TUf4RQ8GYIBCIwD5RkOwj92echkTltUCFsygQ
    b6US4a5WYAg+UNAgSCpzTkj/tGAAtmB4qhN8LHXUOzM0EjChFG/3WJffAoGBAM3d
    Yn9Zq8MjyRViFMgOQzxcv4EfY1tiE2IVJ7skRkI/KBWcpAqg54N3Ft4ih2RzYUob
    e5xPHMu44MqBrDXnk5RGiDI2ph3xvVszTTsFWCtn6PXrh7v8OTYovsiww/aN10/p
    Rn7zz7oSAKUizyU6tdu6xMOW7GE8lsI/S70aycxRAoGAJqdwwyGuKJnAmSSd7M2C
    b2ZYmPsHLpGYGggF0fsYaBorWm0a1qJhrb2p6eNuQToU3XwQPajyghKjeejTdw/F
    5j/S0OSYCRY9OACj7JTqigXkZPUX1YJNZYJjtxGMHS6A9TY1fFg/nV0zEV5PWjOL
    3/8RQvqWvHMFKHBd6FCqNmUCgYAb2/rpaxQwi1Y6G5TeYfe9YnvUGJBUnJgs7Nn8
    nHMZofxluFYGzjGme+ZPV3LlKCwhYEjBJX+rHjDlltjcTqONLGJgET83zDAo+G9a
    LmX5Mc24AhDTYtXHO4peFHXglt9thA8zPQF+l9MYhfZsfl6ABu173p/MpOtuDCzO
    waJPkQKBgGNWifCTY+rfDyzZOO50jefGXALd6rhMscfGED+gwyfFfHQxdiutDLnI
    VPd3tSZu2RU0c3a5wEEFqlJAl07VkVLg96mTKlW7dJzvvS3eqXR3v56f3+MSvTrO
    ZBQeOldZwwPpJEnP4bhDlAOqFtffc3JmvsczOOhYVDkLduuUgVUc
    -----END RSA PRIVATE KEY-----`
}

// pubKeyEntry is test helper (hence t in params) to generate a valid pub Key entry
func pubKeyEntry(t *testing.T, name string) string {
	return `
- name: "` + name + `"
  comment: "generated with: openssl rsa -in private.pem -pubout -outform PEM -out public.pem"
  not_before: 2020-01-01
  not_after: 2022-01-01
  key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArUG/lcVDen29RnHK+f1E
    5UzoXAAMTT5Xatdts7o4+jNwXl7LuYSAnmrl50XjyM5fLCog6G+qLz0L6U07EbXB
    0B/paHuYLlAG9rIYIaAceZYrhMReUSx3DL2yIpawa1YR9QYgyHTY2/3sXF+vd/T7
    JNqBxI/v0vZkhaFCugrWlvAICC1YjQXrjZqRRPl0OsUwZ2kmRvlqvYcVSLEif+uK
    eNMyplThb9CEQZdjjLMSskYzcmGSfPc10ciEDhYR4O5M8vOO5DLeLwj6dw/PTrrs
    lAxrdQqiP4/xyx89ZfFMsxBIBw5JeZ0VnQ46Chr5Dy34A/FacA3Sqb0ZEFkmwCTB
    jwIDAQAB
    -----END PUBLIC KEY-----`
}

func TestBasicViperEmptyConfig(t *testing.T) {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample = []byte(`
JWTPublicKey: []
JWTPrivateKey: []
`)

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlExample)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	arr, err := vc.JWTPublicKeys()
	assert.Error(t, err)
	assert.Equal(t, 0, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Equal(t, JWTPrivateKey{}, pkey)
}

func TestBasicViperNoBootstrap(t *testing.T) {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "public") + `
JWTPrivateKey:
` + privKeyEntry(t, "private", true) + `
`)

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlExample)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	arr, err := vc.JWTPublicKeys()
	assert.Error(t, err)
	assert.Equal(t, 0, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Equal(t, JWTPrivateKey{}, pkey)
}

func TestBasicViperPublicKeys(t *testing.T) {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "generated-for-test-1") + `
` + pubKeyEntry(t, "generated-for-test-2") + `
` + pubKeyEntry(t, "generated-for-test-3") + `
`)

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlExample)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	arr, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(arr))
}

func TestBasicViperPrivateKeysOneEnabled(t *testing.T) {
	var yamlOneEnabled = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "generated-for-test-1", true) + `
` + privKeyEntry(t, "generated-for-test-2", false) + `
`)

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlOneEnabled)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	err := vc.Bootstrap()
	assert.NoError(t, err)

	key, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, key.Key)
	assert.Equal(t, "generated-for-test-1", key.Name)
	assert.NotEqual(t, "the-same-but-copied", key.Name)
	assert.NotEmpty(t, key.CreatedAt)
	assert.NotEmpty(t, key.Author)
}
func TestBasicViperPrivateKeysBothEnabled(t *testing.T) {
	var yamlBothEnabled = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "generated-for-test-1", true) + `
` + privKeyEntry(t, "generated-for-test-2", true) + `
`)

	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlBothEnabled)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	key, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, key.Key)
	assert.Equal(t, "generated-for-test-1", key.Name)
	assert.NotEqual(t, "generated-for-test-2", key.Name)
	assert.NotEmpty(t, key.CreatedAt)
	assert.NotEmpty(t, key.Author)
}

func TestBasicViperPrivateKeysAllDisabled(t *testing.T) {
	var yamlBothDisabled = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "generated-for-test-1", false) + `
` + privKeyEntry(t, "generated-for-test-2", false) + `
`)
	vc := NewViperConfig("test", WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlBothDisabled)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	key, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Nil(t, key.Key)
	assert.Empty(t, key.CreatedAt)
	assert.Empty(t, key.Author)
}

func TestBasicViperMerge(t *testing.T) {
	var yamlPrivate1 = []byte(`
JWTPrivateKey:
` + privKeyEntry(t, "private", true) + `
`)
	var yamlPublic1 = []byte(`
JWTPublicKey:
` + pubKeyEntry(t, "public") + `
`)

	vc := NewViperConfig("public",
		WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlPublic1)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	vc2 := NewViperConfig("private",
		WithConfigFormat("yaml"),
		WithConfigSource(bytes.NewBuffer(yamlPrivate1)),
		WithAutoBootstrap(false),
		WithWatchChanges(false),
	)
	assert.NoError(t, vc2.Bootstrap())

	assert.NoError(t, vc.MergeAndDisableHotReload(vc2))
	assert.Equal(t, "merged-public+private", vc.name)

	arr, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(arr))
	assert.NotEmpty(t, arr[0].Comment)
	assert.NotEmpty(t, arr[0].Key)
	assert.NotEmpty(t, arr[0].NotAfter)
	assert.NotEmpty(t, arr[0].NotBefore)

	key, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, key.Key)
	assert.NotEmpty(t, key.CreatedAt)
	assert.NotEmpty(t, key.Author)
}
