package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cretz/go-safeclient/client"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

var safeClient *client.Client

func TestMain(m *testing.M) {
	// Re-randomize each test
	rand.Seed(time.Now().UTC().UnixNano())
	// If we have a integration.conf.json, use that
	var conf client.Conf
	if _, err := os.Stat("integration.conf.json"); !os.IsNotExist(err) {
		byts, err := ioutil.ReadFile("integration.conf.json")
		if err != nil {
			panic(fmt.Errorf("Unable to read conf: %v", err))
		}
		// JSON unmarshal into the conf
		if err := json.Unmarshal(byts, &conf); err != nil {
			panic(fmt.Errorf("Invalid conf: %v", err))
		}
	}
	safeClient = client.NewClient(conf)
	flag.Parse()
	if testing.Verbose() {
		safeClient.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	// Make sure we are authed
	err := safeClient.EnsureAuthed(client.AuthInfo{
		App: client.AuthAppInfo{
			Name:    "SAFE Client Integration Tests",
			ID:      "go-safeclient-tests80.cretz.github.com",
			Version: "0.0.1",
			Vendor:  "cretz",
		},
		Permissions: []string{client.AuthPermSafeDriveAccess},
	})
	if err != nil {
		panic(fmt.Errorf("Unable to ensure authed: %v", err))
	}

	// Save the config just in case it changed
	persistBytes, err := json.Marshal(safeClient.Conf)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("integration.conf.json", persistBytes, 600); err != nil {
		panic(fmt.Errorf("Unable to write integration.conf.json: %v", err))
	}

	// Now run
	os.Exit(m.Run())
}

// We trust this all runs in the same gorouting
const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var alreadyUsed = map[string]struct{}{}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomName() string {
	for {
		str := randomString(20)
		if _, ok := alreadyUsed[str]; !ok {
			alreadyUsed[str] = struct{}{}
			return str
		}
	}
}

func requireReadCloserEqualsString(t *testing.T, expected string, rc io.ReadCloser) {
	defer rc.Close()
	byts, err := ioutil.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, expected, string(byts))
}
