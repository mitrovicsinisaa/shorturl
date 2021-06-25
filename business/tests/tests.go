// Package tests contains supporting code for running tests.
package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/business/data/schema"
	"github.com/mitrovicsinisaa/shorturl/business/data/user"
	"github.com/mitrovicsinisaa/shorturl/foundation/database"
	"github.com/mitrovicsinisaa/shorturl/foundation/web"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Configuration for running tests.
var (
	dbImage = "postgres:13-alpine"
	dbPort  = "5432"
	dbArgs  = []string{"-e", "POSTGRES_PASSWORD=postgres"}
	AdminID = "4f8bc258-d2d1-11eb-b8bc-0242ac130003"
	UserID  = "4f8bc4f6-d2d1-11eb-b8bc-0242ac130003"
)

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T) (*log.Logger, *sqlx.DB, func()) {

	c := startContainer(t, dbImage, dbPort, dbArgs...)

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Waiting for database to be ready ...")

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		dumpContainerLogs(t, c.ID)
		stopContainer(t, c.ID)
		t.Fatalf("database never ready: %v", pingError)
	}

	if err := schema.Migrate(db); err != nil {
		stopContainer(t, c.ID)
		t.Fatalf("migrating error: %s", err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		stopContainer(t, c.ID)
	}

	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return log, db, teardown
}

// Context returns an app level context for testing
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New().String(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}

// Test owns state for running and shutting down tests.
type Test struct {
	TraceID string
	DB      *sqlx.DB
	Log     *log.Logger
	Auth    *auth.Auth
	KID     string

	t       *testing.T
	cleanup func()
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T) *Test {
	log, db, cleanup := NewUnit(t)

	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	// Create RSA keys to enable authentication in our service.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator using this key lookup function to retrieve
	// the corresponding public key.
	kidID := "73dd3c4e-d417-11eb-b8bc-0242ac130003"
	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case kidID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}

	auth, err := auth.New("RS256", lookup, auth.Keys{kidID: privateKey})
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		TraceID: "00000000-0000-0000-0000-000000000000",
		DB:      db,
		Log:     log,
		Auth:    auth,
		KID:     kidID,
		t:       t,
		cleanup: cleanup,
	}

	return &test
}

// Teardown releases any resources used for the test.
func (test *Test) Teardown() {
	test.cleanup()
}

// Token generates an authenticated token for a user.
func (test *Test) Token(kid string, email, pass string) string {
	u := user.New(test.Log, test.DB)
	claims, err := u.Authenticate(context.Background(), test.TraceID, time.Now(), email, pass)
	if err != nil {
		test.t.Fatal(err)
	}

	token, err := test.Auth.GenerateToken(kid, claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}
