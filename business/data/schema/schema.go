// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {

	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)
	return d.Migrate()
}

var migrations = []darwin.Migration{
	{
		Version:     1.1,
		Description: "Create table users",
		Script: `
CREATE TABLE users (
	user_id			UUID,
	name 			TEXT,
	email 			TEXT UNIQUE,
	roles 			TEXT[],
	password_hash 	TEXT,
	date_created 	TIMESTAMP,
	date_updated	TIMESTAMP,

	PRIMARY KEY (user_id)
);`,
	},
	{
		Version:     1.2,
		Description: "Create table shorturl",
		Script: `
CREATE TABLE shorturls (
	shorturl_id		SERIAL,
	url 			TEXT,
	visits 			INT,
	date_created 	TIMESTAMP,
	date_updated	TIMESTAMP,

	PRIMARY KEY (shorturl_id)
);`,
	},
	{
		Version:     1.3,
		Description: "Create table shorturl",
		Script: `
ALTER TABLE shorturls 
	ALTER COLUMN visits SET DEFAULT 0
;`,
	},
}
