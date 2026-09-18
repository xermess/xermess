package database

// pageRequest is how much of a table to read.
//
// A page is asked for by how many rows and how far in, as the other lists in
// this API are: the panel's table is the same table, and a row that has to be
// read by eye is not one anybody pages through a thousand at a time.
type pageRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// The page a request asks for, held between what is worth fetching and what is
// worth sending.
const (
	defaultLimit = 50
	maxLimit     = 200
)

// page returns the limit and offset to read with, correcting anything the
// request asked for that is not a page.
func (r pageRequest) page() (limit, offset int) {
	limit = r.Limit
	switch {
	case limit <= 0:
		limit = defaultLimit
	case limit > maxLimit:
		limit = maxLimit
	}

	offset = r.Offset
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
