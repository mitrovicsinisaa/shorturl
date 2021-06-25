package shorturl

import (
	"time"
)

// Info represents an individual Shorturl.
type Info struct {
	ID          int       `db:"shorturl_id" json:"id"`
	URL         string    `db:"url" json:"url"`
	Visits      int       `db:"visits" json:"visits"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
}

// NewShorturl contains information needed to create a new Shorturl.
type NewShorturl struct {
	URL string `json:"url" validate:"required"`
}

// CreateShorturl contains information needed after create a new Shorturl.
type CreateShorturl struct {
	ID          int       `db:"shorturl_id" json:"id"`
	URL         string    `db:"url" json:"url"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
}

// ShorturlVisits contains information about number of visits for shorturl.
type ShorturlVisits struct {
	Visits int `db:"visits" json:"visits"`
}
