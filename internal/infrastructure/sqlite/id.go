package sqlite

import "github.com/oklog/ulid/v2"

func newID(prefix string) string {
	return prefix + "_" + ulid.Make().String()
}
