package fields

// targetType is what these rows are called in the activity log.
const targetType = "user_field"

// rulesRequest is what an update sends: what a field expects of its values.
// Its name and its type are not here, because neither may change once records
// hold values under them.
type rulesRequest struct {
	Label      string   `json:"label" validate:"max=100"`
	IsRequired bool     `json:"is_required"`
	IsUnique   bool     `json:"is_unique"`
	Min        *float64 `json:"min"`
	Max        *float64 `json:"max"`
	StartsWith string   `json:"starts_with" validate:"max=64"`
}

// fieldRequest is what a create sends: the same rules, plus what the field is.
type fieldRequest struct {
	rulesRequest

	Name string `json:"name" validate:"required,column"`
	Type string `json:"type" validate:"required,fieldtype"`
}
