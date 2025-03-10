package secrets

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var drivertype = "file"

var opts map[string]string

func setup() (*SecretsManager, string, error) {
	testpath, err := os.MkdirTemp("", "secretsdata")
	if err != nil {
		return nil, "", err
	}
	manager, err := NewManager(testpath)
	opts = map[string]string{"path": testpath}
	return manager, testpath, err
}

func cleanup(testpath string) {
	os.RemoveAll(testpath)
}

func TestAddSecretAndLookupData(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	metaData := make(map[string]string)
	metaData["immutable"] = "true"

	labels := make(map[string]string)
	labels["foo"] = "bar"
	labels["another"] = "label"

	storeOpts := StoreOptions{
		DriverOpts: opts,
		Metadata:   metaData,
		Labels:     labels,
	}

	id1, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.lookupSecret("mysecret")
	require.NoError(t, err)

	s, data, err := manager.LookupSecretData("mysecret")
	require.NoError(t, err)
	if !bytes.Equal(data, []byte("mydata")) {
		t.Errorf("error: secret data not equal")
	}
	if val, ok := s.Metadata["immutable"]; !ok || val != "true" {
		t.Errorf("error: no metadata")
	}
	if val, ok := s.Labels["foo"]; !ok || val != "bar" {
		t.Errorf("error: label incorrect")
	}
	if len(s.Labels) != 2 {
		t.Errorf("error: incorrect number of labels")
	}
	if s.CreatedAt != s.UpdatedAt {
		t.Errorf("error: secret CreatedAt should equal UpdatedAt when first created")
	}

	_, err = manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)

	storeOpts.Replace = true
	id2, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)
	if id1 == id2 {
		t.Errorf("error: secret id after Replace should be different")
	}

	s, _, err = manager.LookupSecretData("mysecret")
	require.NoError(t, err)
	if s.CreatedAt == s.UpdatedAt {
		t.Errorf("error: secret CreatedAt should not equal UpdatedAt after a Replace")
	}

	_, _, err = manager.LookupSecretData(id2)
	require.NoError(t, err)

	_, _, err = manager.LookupSecretData(id1)
	require.Error(t, err)
}

func TestAddSecretName(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	// test one char secret name
	_, err = manager.Store("a", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.lookupSecret("a")
	require.NoError(t, err)

	// name too short
	_, err = manager.Store("", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	// name too long
	_, err = manager.Store("uatqsbssrapurkuqoapubpifvsrissslzjehalxcesbhpxcvhsozlptrmngrivaizuatqsbssrapurkuqoapubpifvsrissslzjehalxcesbhpxcvhsozlptrmngrivaizuatqsbssrapurkuqoapubpifvsrissslzjehalxcesbhpxcvhsozlptrmngrivaizuatqsbssrapurkuqoapubpifvsrissslzjehalxcesbhpxcvhsozlptrabd", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	// invalid chars
	_, err = manager.Store("??", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	_, err = manager.Store("-a", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	_, err = manager.Store("a-", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	_, err = manager.Store(".a", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
	_, err = manager.Store("a.", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
}

func TestAddMultipleSecrets(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	id, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	id2, err := manager.Store("mysecret2", []byte("mydata2"), drivertype, storeOpts)
	require.NoError(t, err)

	secrets, err := manager.List()
	require.NoError(t, err)
	require.Len(t, secrets, 2)

	_, err = manager.lookupSecret("mysecret")
	require.NoError(t, err)

	_, err = manager.lookupSecret("mysecret2")
	require.NoError(t, err)

	_, data, err := manager.LookupSecretData(id)
	require.NoError(t, err)
	if !bytes.Equal(data, []byte("mydata")) {
		t.Errorf("error: secret data not equal")
	}

	_, data2, err := manager.LookupSecretData(id2)
	require.NoError(t, err)
	if !bytes.Equal(data2, []byte("mydata2")) {
		t.Errorf("error: secret data not equal")
	}
}

func TestAddSecretDupName(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	_, err = manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.Error(t, err)
}

func TestAddSecretPrefix(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	// If the randomly generated secret id is something like "abcdeiuoergnadufigh"
	// we should still allow someone to store a secret with the name "abcd" or "a"
	secretID, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.Store(secretID[0:5], []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)
}

func TestRemoveSecret(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	_, err = manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.lookupSecret("mysecret")
	require.NoError(t, err)

	_, err = manager.Delete("mysecret")
	require.NoError(t, err)

	_, err = manager.lookupSecret("mysecret")
	require.Error(t, err)

	_, _, err = manager.LookupSecretData("mysecret")
	require.Error(t, err)
}

func TestRemoveSecretNoExist(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	_, err = manager.Delete("mysecret")
	require.Error(t, err)
}

func TestLookupAllSecrets(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	id, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	// inspect using secret name
	lookup, err := manager.Lookup("mysecret")
	require.NoError(t, err)
	require.Equal(t, lookup.ID, id)
}

func TestInspectSecretId(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	id, err := manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)

	_, err = manager.lookupSecret("mysecret")
	require.NoError(t, err)

	// inspect using secret id
	lookup, err := manager.Lookup(id)
	require.NoError(t, err)
	require.Equal(t, lookup.ID, id)

	// inspect using id prefix
	short := id[0:5]
	lookupshort, err := manager.Lookup(short)
	require.NoError(t, err)
	require.Equal(t, lookupshort.ID, id)
}

func TestInspectSecretBogus(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	_, err = manager.Lookup("bogus")
	require.Error(t, err)
}

func TestSecretList(t *testing.T) {
	manager, testpath, err := setup()
	require.NoError(t, err)
	defer cleanup(testpath)

	storeOpts := StoreOptions{
		DriverOpts: opts,
	}

	_, err = manager.Store("mysecret", []byte("mydata"), drivertype, storeOpts)
	require.NoError(t, err)
	_, err = manager.Store("mysecret2", []byte("mydata2"), drivertype, storeOpts)
	require.NoError(t, err)

	allSecrets, err := manager.List()
	require.NoError(t, err)
	require.Len(t, allSecrets, 2)
}
