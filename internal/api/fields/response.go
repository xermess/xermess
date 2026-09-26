package fields

import "loginer/internal/model"

// fieldResponse is one field as the panel sees it.
//
// It is built by hand rather than returning the model because the two kinds
// of field are not the same shape: an additional one is a row and has an id
// and timestamps, while a built-in one is a column of the record and has
// none. Sending a row's fields for something that is not a row would be
// saying it has an id of all zeros, which is not true.
type fieldResponse struct {
	ID string `json:"id,omitempty"`

	Name  string          `json:"name"`
	Label string          `json:"label"`
	Type  model.FieldType `json:"type"`

	Required   bool     `json:"required"`
	Unique     bool     `json:"unique"`
	Min        *float64 `json:"min"`
	Max        *float64 `json:"max"`
	StartsWith string   `json:"starts_with"`

	Position int  `json:"position"`
	Builtin  bool `json:"builtin"`
}

func newFieldResponse(field model.UserField) fieldResponse {
	out := fieldResponse{
		Name:       field.Name,
		Label:      field.Label,
		Type:       field.Type,
		Required:   field.Required,
		Unique:     field.Unique,
		Min:        field.Min,
		Max:        field.Max,
		StartsWith: field.StartsWith,
		Position:   field.Position,
		Builtin:    field.Builtin,
	}

	if !field.Builtin {
		out.ID = field.ID.String()
	}

	return out
}

// listResponse is every field a user record has, and the types a new one may
// have. Both come together because the panel's form needs the second to offer
// the first.
//
// The built-in fields come first and are marked as such: the panel draws its
// table and its form from this one list, and uses the mark to know which of
// them it may offer to change.
type listResponse struct {
	Fields []fieldResponse   `json:"fields"`
	Types  []model.FieldType `json:"types"`
}

func newListResponse(added []model.UserField) listResponse {
	builtin := model.BuiltinFields()
	fields := make([]fieldResponse, 0, len(builtin)+len(added))

	for _, field := range builtin {
		fields = append(fields, newFieldResponse(field))
	}

	for _, field := range added {
		fields = append(fields, newFieldResponse(field))
	}

	return listResponse{Fields: fields, Types: model.FieldTypes}
}
