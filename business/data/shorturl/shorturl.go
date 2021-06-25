// Package shorturl contains shorturl related CRUD functionality.
package shorturl

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/foundation/database"
	"github.com/pkg/errors"
)

var (
	// ErrNotFound is used when a specific Shorturl is requested but does not exists.
	ErrNotFound = errors.New("not found")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbiden.
	ErrForbidden = errors.New("action is not allowed")
)

// Shorturl manages the set of API's for shorturl access.
type Shorturl struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a shorturl for api access.
func New(log *log.Logger, db *sqlx.DB) Shorturl {
	return Shorturl{
		log: log,
		db:  db,
	}
}

// Create inserts a new shorturl into the database.
func (su Shorturl) Create(ctx context.Context, traceID string, nsu NewShorturl, now time.Time) (CreateShorturl, error) {

	shorturl := CreateShorturl{
		URL:         nsu.URL,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
	INSERT INTO shorturls
		(url, date_created, date_updated)
	VALUES
		($1, $2, $3)
		RETURNING shorturl_id;`

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.Create",
		database.Log(q, shorturl.URL, shorturl.DateCreated, shorturl.DateUpdated))

	if err := su.db.GetContext(ctx, &shorturl, q, shorturl.URL, shorturl.DateCreated, shorturl.DateUpdated); err != nil {
		return CreateShorturl{}, errors.Wrap(err, "inserting shorturl")
	}

	su.log.Printf("%s : %s : query : %v", traceID, "shorturlID", nsu.URL)

	return shorturl, nil
}

// Delete removes a shorturl from the database.
func (su Shorturl) Delete(ctx context.Context, traceID string, claims auth.Claims, shorturl_id int) error {

	const q = `
	DELETE FROM
		shorturls
	WHERE
		shorturl_id = $1`

	shorturlID := 1

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.Delete",
		database.Log(q, shorturl_id))

	if _, err := su.db.ExecContext(ctx, q, shorturlID); err != nil {
		return errors.Wrapf(err, "deleting shorturl %s", shorturl_id)
	}

	return nil
}

// Query retrieves a list of existing shorturls from the database.
func (su Shorturl) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {

	const q = `
	SELECT
		*
	FROM
		shorturls
	ORDER BY
		shorturl_id
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.Query",
		database.Log(q, pageNumber, rowsPerPage))

	shorturls := []Info{}
	if err := su.db.SelectContext(ctx, &shorturls, q, pageNumber, rowsPerPage); err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "selecting shorturl")
	}

	return shorturls, nil
}

// QueryByURL gets the specified shorturl from the database.
func (su Shorturl) QueryByID(ctx context.Context, traceID string, shorturl_id int) (Info, error) {

	const q = `
	SELECT
		*
	FROM
		shorturls
	WHERE 
		shorturl_id = $1`

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.QueryByID",
		database.Log(q, shorturl_id))

	var shorturl Info
	if err := su.db.GetContext(ctx, &shorturl, q, shorturl_id); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting shorturl %q", shorturl_id)
	}

	const uq = `
	UPDATE
		shorturls
	SET 
		"visits" = $2
	WHERE
		shorturl_id = $1`

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.QueryByID",
		database.Log(uq, shorturl_id, shorturl.Visits+1))

	if _, err := su.db.ExecContext(ctx, uq, shorturl_id, shorturl.Visits+1); err != nil {
		return Info{}, errors.Wrapf(err, "updating visits shorturl %q", shorturl_id)
	}

	return shorturl, nil
}

// QueryVisitation gets the number of visits for specified shorturl from the database.
func (su Shorturl) QueryVisitation(ctx context.Context, traceID string, claims auth.Claims, shorturl_id int) (ShorturlVisits, error) {

	const q = `
	SELECT
		visits
	FROM
		shorturls
	WHERE 
		shorturl_id = $1`

	su.log.Printf("%s : %s : query : %s", traceID, "shorturl.queryVisitation",
		database.Log(q, shorturl_id))

	var visits ShorturlVisits
	if err := su.db.GetContext(ctx, &visits, q, shorturl_id); err != nil {
		if err == sql.ErrNoRows {
			return ShorturlVisits{}, ErrNotFound
		}
		return ShorturlVisits{}, errors.Wrapf(err, "selecting shorturl visits %q", shorturl_id)
	}

	return visits, nil
}
