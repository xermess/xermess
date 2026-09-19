package fields

import (
	"net/http"
	"regexp"
	"strings"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// column is what a field may be called: the same shape as a column name,
// because that is what it stands in for.
var column = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// The two rules a request here is held to that the validator cannot know by
// itself. They are registered where they are meant, so the rule and the thing
// it describes stay together.
func init() {
	validate.Register("column", column.MatchString)

	validate.Register("fieldtype", func(value string) bool {
		return model.FieldType(value).Valid()
	})
}

// newField turns a create request into the field it describes, or says what
// is wrong with it.
func (r *fieldRequest) newField() (*model.UserField, error) {
	r.Name = strings.ToLower(strings.TrimSpace(r.Name))
	r.Type = strings.TrimSpace(r.Type)

	if err := validate.Struct(r); err != nil {
		return nil, err
	}

	if model.IsBuiltinField(r.Name) {
		return nil, respond.Fault{
			Status: http.StatusConflict,
			Message: r.Name + " is a built-in field: every record has one already, " +
				"and two fields with one name would be two places to look",
		}
	}

	field := &model.UserField{Name: r.Name, Type: model.FieldType(r.Type)}
	if err := r.rulesRequest.applyTo(field); err != nil {
		return nil, err
	}

	return field, nil
}

// applyTo copies the rules from a request onto a field and checks that the
// field can keep them — a bool with a length, or a largest value below its
// smallest, is a mistake in the panel rather than in someone's record.
//
// The name and the type are not touched: an update may not change either.
func (r *rulesRequest) applyTo(field *model.UserField) error {
	r.Label = strings.TrimSpace(r.Label)
	r.StartsWith = strings.TrimSpace(r.StartsWith)

	if err := validate.Struct(r); err != nil {
		return err
	}

	label := r.Label
	if label == "" {
		label = field.Name
	}

	field.Label = label
	field.Required = r.Required
	field.Unique = r.Unique
	field.Min = r.Min
	field.Max = r.Max
	field.StartsWith = r.StartsWith

	// What one rule means next to another is the model's to say: it owns what
	// a field of each type can carry.
	if err := field.Validate(); err != nil {
		return respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
	}

	return nil
}
