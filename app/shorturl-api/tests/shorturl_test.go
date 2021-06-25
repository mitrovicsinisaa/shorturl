package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mitrovicsinisaa/shorturl/app/shorturl-api/handlers"
	"github.com/mitrovicsinisaa/shorturl/business/data/shorturl"
	"github.com/mitrovicsinisaa/shorturl/business/tests"
	"github.com/mitrovicsinisaa/shorturl/foundation/web"
)

// ShorturlTests holds methods for each shorturl subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type ShorturlTests struct {
	app        http.Handler
	kid        string
	adminToken string
}

// TestShorturl is the entry point for testing shorturl API functions.
func TestShorturl(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := ShorturlTests{
		app:        handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		adminToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("postShorturl400", tests.postShorturl400)
	// t.Run("postUser401", tests.postUser401)
	// t.Run("postUser403", tests.postUser403)
	// t.Run("getUser400", tests.getUser400)
	// t.Run("getUser403", tests.getUser403)
	// t.Run("getUser404", tests.getUser404)
	// t.Run("deleteUserNotFound", tests.deleteUserNotFound)
	// t.Run("putUser404", tests.putUser404)
	// t.Run("crudUsers", tests.crudUser)
}

// postShorturl400 validates a user can't be created with the endpoint
// unless a valid user document is submitted.
func (st *ShorturlTests) postShorturl400(t *testing.T) {
	body, err := json.Marshal(&shorturl.NewShorturl{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/shorturl", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+st.adminToken)
	st.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new shorturl can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete shorturl value.", testID)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", tests.Success, testID)

			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type.", tests.Success, testID)
		}
	}
}
