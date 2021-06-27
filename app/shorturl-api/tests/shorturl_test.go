package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mitrovicsinisaa/shorturl/app/shorturl-api/handlers"
	"github.com/mitrovicsinisaa/shorturl/business/base62"
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
	t.Run("getShorturl404", tests.getShorturl404)
	t.Run("getShorturl401", tests.getShorturl401)
	t.Run("successShorturlActions", tests.successShorturlActions)
}

// postShorturl400 validates a shorturl can't be created with the endpoint
// unless a valid user document is submitted.
func (st *ShorturlTests) postShorturl400(t *testing.T) {
	body, err := json.Marshal(&shorturl.NewShorturl{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/shorturl", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

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

// getShorturl404 validates a shorturl request for a bad url.
func (st *ShorturlTests) getShorturl404(t *testing.T) {
	url := "owna"

	r := httptest.NewRequest(http.MethodGet, "/"+url, nil)
	w := httptest.NewRecorder()

	st.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a shorturl with a bad url.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen getting the shorturl %s.", testID, url)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)
		}
	}
}

// getShorturl401 validates a shorturl request for a unauthorized request.
func (st *ShorturlTests) getShorturl401(t *testing.T) {
	page := "0"
	rows := "1"

	r := httptest.NewRequest(http.MethodGet, "/api/shorturl/"+page+"/"+rows, nil)
	w := httptest.NewRecorder()

	st.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a shorturls with a unauthorized request.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen getting the new shorturls page %s and count %s.", testID, page, rows)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// successShorturlActions performs a complete test of success actions against
// the shorturl api.
func (st *ShorturlTests) successShorturlActions(t *testing.T) {
	st.postShorturl201(t)
	surl := st.getShorturl200(t)
	st.getShorturlVisits200(t, surl.ID)
	defer st.deleteShorturl204(t, surl.ID)
}

// postShorturl201 validates a shorturl can be created with the endpoint.
func (st *ShorturlTests) postShorturl201(t *testing.T) {
	ns := shorturl.NewShorturl{
		URL: "https://www.google.com/",
	}

	body, err := json.Marshal(&ns)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/shorturl", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+st.adminToken)
	st.app.ServeHTTP(w, r)

	var got struct {
		ShortUrl string `json:"shorturl"`
	}
	t.Log("Given the need to create a new shorturl with the shorturl endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared shorturl value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}
		}
	}
}

// getShorturl200 validates that list of shortulrls can be fetched.
func (st *ShorturlTests) getShorturl200(t *testing.T) shorturl.Info {
	page := "0"
	rows := "1"

	r := httptest.NewRequest(http.MethodGet, "/api/shorturl/"+page+"/"+rows, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+st.adminToken)
	st.app.ServeHTTP(w, r)

	var got []shorturl.Info
	t.Log("Given the need to validate getting a list of shorturls.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen getting the list of shorturls for page %s and count %s.", testID, page, rows)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}
		}
	}

	return got[0]
}

// getShorturl200 validates that shortulrl visit count can be fetched.
func (st *ShorturlTests) getShorturlVisits200(t *testing.T, ID int) {
	url := base62.Encode(ID)

	rv := httptest.NewRequest(http.MethodGet, "/"+url, nil)
	wv := httptest.NewRecorder()
	st.app.ServeHTTP(wv, rv)
	st.app.ServeHTTP(wv, rv)
	st.app.ServeHTTP(wv, rv)

	r := httptest.NewRequest(http.MethodGet, "/api/shorturl/"+url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+st.adminToken)
	st.app.ServeHTTP(w, r)

	var got shorturl.ShorturlVisits
	t.Log("Given the need to validate getting a count of shorturl visits.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen getting the shorturl visits.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if diff := cmp.Diff(got.Visits, 3); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// deleteShorturl204 validates deleting a shorturl that does exist.
func (st *ShorturlTests) deleteShorturl204(t *testing.T, ID int) {
	url := base62.Encode(ID)
	r := httptest.NewRequest(http.MethodDelete, "/api/shorturl/"+url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+st.adminToken)
	st.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a shorturl that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new shorturl %s.", testID, url)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}
