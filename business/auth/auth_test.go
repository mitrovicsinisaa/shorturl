package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitrovicsinisaa/shorturl/business/auth"
)

// Success and failure markers.

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			// Generate a new private key.
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				log.Fatalln(err)
			}

			// The key id we are stating represents the public key in the
			// public key store.
			const keyID = "f0fb59d8-cd30-11eb-b8bc-0242ac130003"
			lookup := func(kid string) (*rsa.PublicKey, error) {
				switch kid {
				case keyID:
					return &privateKey.PublicKey, nil
				}
				return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
			}

			a, err := auth.New("RS256", lookup, auth.Keys{keyID: privateKey})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:Should be able to create an authenticator: %v", failure, testID, err)
			}
			t.Logf("\t%s\tTest %d:Should be able to create an authenticator.", success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "b02f5bd7-3640-4328-a467-44ef27e23eee",
					Audience:  "Students",
					ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := a.GenerateToken(keyID, claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:Should be able to generate a JWT: %v", failure, testID, err)
			}
			t.Logf("\t%s\tTest %d:Should be able to generate a JWT.", success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:Should be able to parse the claims: %v", failure, testID, err)
			}
			t.Logf("\t%s\tTest %d:Should be able to parse the claims.", success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t\tTest %d:\texp %d", testID, exp)
				t.Logf("\t\tTest %d:\tgot %d", testID, got)
				t.Fatalf("\t%s\tTest %d:Should have the expected number of roles: %v", failure, testID, err)
			}
			t.Logf("\t%s\tTest %d:Should have the expected number of roles.", success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t\tTest %d:\texp %v", testID, exp)
				t.Logf("\t\tTest %d:\tgot %v", testID, got)
				t.Fatalf("\t%s\tTest %d:Should have the expected roles: %v", failure, testID, err)
			}
			t.Logf("\t%s\tTest %d:Should have the expected roles.", success, testID)
		}
	}
}
