package shorturl_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/business/data/shorturl"
	"github.com/mitrovicsinisaa/shorturl/business/tests"
	"github.com/pkg/errors"
)

func TestShorturl(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	su := shorturl.New(log, db)

	t.Log("Given the need to work with Shorturl records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Shorturl.", testID)
		{
			ctx := tests.Context()
			now := time.Date(2021, time.June, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			ns := shorturl.NewShorturl{
				URL: "https://github.com/mitrovicsinisaa/shorturl",
			}

			surl, err := su.Create(ctx, traceID, ns, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create shorturl : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create shorturl.", tests.Success, testID)

			_, err = su.QueryByID(ctx, traceID, surl.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve shorturl by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve shorturl by ID.", tests.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   strconv.Itoa(surl.ID),
					ExpiresAt: now.Add(time.Hour).Unix(),
					IssuedAt:  now.Unix(),
				},
				Roles: []string{auth.RoleUser},
			}

			if err := su.Delete(ctx, traceID, claims, surl.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete shorturl : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete shorturl.", tests.Success, testID)

			_, err = su.QueryByID(ctx, traceID, surl.ID)
			if errors.Cause(err) != shorturl.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve shorturl : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve shorturl.", tests.Success, testID)
		}
	}
}
