package main

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// errInvalidUUID means a text is not a UUID. It usually means the caller
// sent a bad id, not that something broke on our side.
var errInvalidUUID = errors.New("not a valid uuid")

// parseUUID turns a UUID in text form into the type the queries use.
func parseUUID(text string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(text)
	if err != nil {
		// %w wraps errInvalidUUID, so callers can find it with errors.Is.
		return pgtype.UUID{}, fmt.Errorf("%w: %q", errInvalidUUID, text)
	}
	return id, nil
}
