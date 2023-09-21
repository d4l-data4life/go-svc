package dynamic_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
)

// TestViperConfigCopying ensures that ViperConfing object works correctly when copied in various ways (as long as 'go vet' allows)
func TestViperConfigCopying(t *testing.T) {
	// create config.yaml with 2 keys: A0, A1
	tmpDir := generateAndWriteConfig(t, 2, "A", "")
	defer func() {
		os.RemoveAll(tmpDir) // clean up the temp dir
	}()

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigFileName("config"),
		dynamic.WithConfigFilePaths(tmpDir),
		dynamic.WithAutoBootstrap(true),
		dynamic.WithWatchChanges(true),
	)
	if vc.Error != nil {
		t.Fatal(vc.Error, "unable to bootstrap VC")
	}

	arr, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "A0", pkey.Name)

	// now get an interface
	type vcInterface interface {
		JWTPublicKeys() ([]dynamic.JWTPublicKey, error)
		JWTPrivateKey() (dynamic.JWTPrivateKey, error)
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
	pointerFun := func(vc *dynamic.ViperConfig, t *testing.T, numPub int, privName, testName string) {
		arr, err := vc.JWTPublicKeys()
		assert.NoError(t, err, testName)
		assert.Equal(t, numPub, len(arr), testName)

		pkey, err := vc.JWTPrivateKey()
		assert.NoError(t, err, testName)
		assert.Equal(t, privName, pkey.Name, testName)
	}

	interfacerFun(vc, t, 2, "A0", "phase1-interfacer")
	pointerFun(vc, t, 2, "A0", "phase1-pointer")

	// now simulate hot-reload of the config
	// KEY ROTATION ON DISK IS HAPPENING HERE
	// replace 4 key entries named `pre` with only one key named `post`
	tmpDir = generateAndWriteConfig(t, 1, "B", tmpDir)
	// wait a bit for the filesystem to catch the changes in the config files and notify Viper
	<-time.After(500 * time.Millisecond)

	// all forms should catch the new changes
	interfacerFun(vc, t, 1, "B0", "phase2-interfacer")
	pointerFun(vc, t, 1, "B0", "phase2-pointer")
}

// TestViperConfigHotReloadAfterMerge ensures that ViperConfing (VC) object works (in)correctly after merging with another VC object
func TestViperConfigHotReloadAfterMerge(t *testing.T) {
	priv1, pub1 := keyEntries(t, "one", true)
	priv2, pub2 := keyEntries(t, "two", false)
	var yamlExample1 = []byte(publicKeyYaml(pub1, pub2))    // only public keys
	var yamlExample2 = []byte(privateKeyYaml(priv1, priv2)) // only private keys - we will merge both later

	// prepare two separate temporary dirs
	tmpDir1, err := os.MkdirTemp("", "unittest1")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created dir %s", tmpDir1)
	defer func() {
		t.Logf("deleting dir %s", tmpDir1)
		os.RemoveAll(tmpDir1) // clean up
	}()
	tmpDir2, err := os.MkdirTemp("", "unittest2")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created dir %s", tmpDir2)
	defer func() {
		t.Logf("deleting dir %s", tmpDir2)
		os.RemoveAll(tmpDir2) // clean up
	}()

	// Write to config.yaml files to separate dirs
	if err := os.WriteFile(filepath.Join(tmpDir1, "config.yaml"), yamlExample1, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir1, "config.yaml"))

	if err := os.WriteFile(filepath.Join(tmpDir2, "config.yaml"), yamlExample2, 0666); err != nil {
		t.Fatal(err)
	}
	t.Logf("written file %s", filepath.Join(tmpDir2, "config.yaml"))

	// build 2 separate vc objects
	vc1 := dynamic.NewViperConfig("test1", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigFileName("config"),
		dynamic.WithConfigFilePaths(tmpDir1),
		dynamic.WithAutoBootstrap(true),
		dynamic.WithWatchChanges(true),
	)
	if vc1.Error != nil {
		t.Fatal(vc1.Error, "unable to bootstrap VC1")
	}
	vc2 := dynamic.NewViperConfig("test2", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigFileName("config"),
		dynamic.WithConfigFilePaths(tmpDir2),
		dynamic.WithAutoBootstrap(true),
		dynamic.WithWatchChanges(true),
	)
	if vc2.Error != nil {
		t.Fatal(vc2.Error, "unable to bootstrap VC2")
	}
	// sanity checks
	arr, err := vc1.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))

	pkey, err := vc2.JWTPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "one", pkey.Name)

	// now merge the two VC into 1
	err = vc1.MergeAndDisableHotReload(vc2)
	assert.NoError(t, err)
	// vc1 consits now of vc1+vc2
	assert.Equal(t, vc1.Name(), "merged-test1+test2")
	// now, the bound between merged vc1 and config file 2 is lost
	// updating config.yaml in tmpDir2 will NOT lead to correct hot-reloading of merged config, however, the config from tmpDir1 should be updated

	// now simulate hot-reload of the config 1 and 2
	privA, pubA := keyEntries(t, "A", true)
	privB, pubB := keyEntries(t, "B", false)
	_, pubC := keyEntries(t, "C", false)
	var yamlExample1B = []byte(publicKeyYaml(pubA, pubB, pubC))

	if err := os.WriteFile(filepath.Join(tmpDir1, "config.yaml"), yamlExample1B, 0666); err != nil {
		t.Fatal(err)
	}
	var yamlExample2B = []byte(privateKeyYaml(privA, privB))

	if err := os.WriteFile(filepath.Join(tmpDir2, "config.yaml"), yamlExample2B, 0666); err != nil {
		t.Fatal(err)
	}
	// wait a bit for the filesystem to catch the changes in the config files and notify Viper
	<-time.After(50 * time.Millisecond)

	arr, err = vc1.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(arr))
	assert.Equal(t, "A", arr[0].Name)
	assert.Equal(t, "B", arr[1].Name)
	assert.Equal(t, "C", arr[2].Name)

	pkey, err = vc1.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotEqual(t, "A", pkey.Name) // We all would like that this would work this way ;(
	assert.Equal(t, "one", pkey.Name)  // instead, the hot-reload for the vc2 is ignored, and we see the old value here

}

func TestBasicViperEmptyConfig(t *testing.T) {
	var yamlExample = []byte(configYaml(publicKeyYaml(), privateKeyYaml())) // two empty entries

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlExample)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	arr, err := vc.JWTPublicKeys()
	assert.Error(t, err)
	assert.Equal(t, 0, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Equal(t, dynamic.JWTPrivateKey{}, pkey)
}

func TestBasicViperNoBootstrap(t *testing.T) {
	priv1, pub1 := keyEntries(t, "one", true)
	var yamlExample = []byte(configYaml(publicKeyYaml(pub1), privateKeyYaml(priv1)))

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlExample)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	arr, err := vc.JWTPublicKeys()
	assert.Error(t, err)
	assert.Equal(t, 0, len(arr))

	pkey, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Equal(t, dynamic.JWTPrivateKey{}, pkey)
}

func TestBasicViperPublicKeys(t *testing.T) {
	_, pub1 := keyEntries(t, "one", true)
	_, pub2 := keyEntries(t, "two", true)
	_, pub3 := keyEntries(t, "three", true)
	// important - watch indentation here! this must produce valid yaml
	yamlExample := publicKeyYaml(pub1, pub2, pub3)

	t.Logf("%s", yamlExample)

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer([]byte(yamlExample))),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	arr, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(arr))
}

func TestBasicViperPrivateKeysOneEnabled(t *testing.T) {
	priv1, _ := keyEntries(t, "one", true)
	priv2, _ := keyEntries(t, "two", false)
	var yamlOneEnabled = []byte(privateKeyYaml(priv1, priv2))

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlOneEnabled)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	err := vc.Bootstrap()
	assert.NoError(t, err)

	key, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, key.Key)
	assert.Equal(t, "one", key.Name)
	assert.NotEqual(t, "two", key.Name)
	assert.NotEmpty(t, key.CreatedAt)
	assert.NotEmpty(t, key.Author)
}

func TestBasicViperPrivateKeysBothEnabled(t *testing.T) {
	priv1, _ := keyEntries(t, "one", true)
	priv2, _ := keyEntries(t, "two", true)
	var yamlBothEnabled = []byte(privateKeyYaml(priv1, priv2))

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlBothEnabled)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	key, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, key.Key)
	assert.Equal(t, "one", key.Name)
	assert.NotEqual(t, "two", key.Name)
	assert.NotEmpty(t, key.CreatedAt)
	assert.NotEmpty(t, key.Author)
}

func TestBasicViperPrivateKeysAllDisabled(t *testing.T) {
	priv1, _ := keyEntries(t, "one", false)
	priv2, _ := keyEntries(t, "two", false)
	var yamlBothDisabled = []byte(privateKeyYaml(priv1, priv2))

	vc := dynamic.NewViperConfig("test", dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlBothDisabled)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	key, err := vc.JWTPrivateKey()
	assert.Error(t, err)
	assert.Nil(t, key.Key)
	assert.Empty(t, key.CreatedAt)
	assert.Empty(t, key.Author)
}

func TestBasicViperMerge(t *testing.T) {
	priv1, pub1 := keyEntries(t, "one", true)

	var yamlPrivate1 = []byte(privateKeyYaml(priv1))
	var yamlPublic1 = []byte(publicKeyYaml(pub1))

	vc := dynamic.NewViperConfig("public",
		dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlPublic1)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc.Bootstrap())

	vc2 := dynamic.NewViperConfig("private",
		dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlPrivate1)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	assert.NoError(t, vc2.Bootstrap())

	assert.NoError(t, vc.MergeAndDisableHotReload(vc2))
	assert.Equal(t, "merged-public+private", vc.Name())

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
