package database

import "xermess/internal/store"

// tablesResponse is what the list endpoint answers with.
type tablesResponse struct {
	Tables []store.Table `json:"tables"`
}

func newTablesResponse(tables []store.Table) tablesResponse {
	if tables == nil {
		tables = []store.Table{}
	}

	return tablesResponse{Tables: tables}
}

// tableResponse is one table: what its columns are, and the page of rows that
// was asked for.
//
// The rows are maps rather than a fixed shape, because the shape is the
// table's. The columns come with them so the panel draws the same columns in
// the same order whether or not the page it fetched happens to have a value
// in each of them.
type tableResponse struct {
	Table   string           `json:"table"`
	Columns []store.Column   `json:"columns"`
	Rows    []map[string]any `json:"rows"`

	// Total is how many rows the table has, which is what the page is a page
	// of.
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

func newTableResponse(
	name string,
	columns []store.Column,
	rows []map[string]any,
	total int64,
	limit, offset int,
) tableResponse {
	if rows == nil {
		rows = []map[string]any{}
	}

	return tableResponse{
		Table:   name,
		Columns: columns,
		Rows:    rows,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}
