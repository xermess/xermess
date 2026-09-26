package users

import (
	"context"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/google/uuid"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
	"loginer/internal/store"
)

// minPasswordLength is the shortest password an administrator may give a
// user. Length is all that is asked of it, as for the super admin's.
const minPasswordLength = 8

// validate checks the built-in fields, which are the record's own columns.
// What the additional ones have to keep is checked by normalise below,
// against the definitions in user_fields rather than against a tag.
//
// `creating` says whether this makes a new user, which is the one time a
// password has to be given.
func (r *userRequest) validate(creating bool) error {
	r.clean()

	if err := validate.Struct(r); err != nil {
		return err
	}

	switch {
	case r.Password == "" && creating:
		return badRequest("password is required")
	case r.Password != "" && utf8.RuneCountInString(r.Password) < minPasswordLength:
		return badRequest(fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}

	return nil
}

func badRequest(message string) error {
	return respond.Fault{Status: http.StatusBadRequest, Message: message}
}

// normalise checks the submitted values against the field definitions and
// returns what should be stored: the rules a field carries — its type, its
// bounds, its prefix, whether it has to be unique — are all applied here.
//
// `self` is the user being written, so a unique field does not find the
// record's own value and call it a clash. It is uuid.Nil when creating one.
func normalise(
	ctx context.Context,
	st *store.Store,
	fields []model.UserField,
	values map[string]any,
	self uuid.UUID,
) (map[string]any, error) {
	data := make(map[string]any, len(fields))

	for _, field := range fields {
		value, err := field.Normalise(values[field.Name])
		if err != nil {
			return nil, respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
		}

		// A field with no value is left out rather than stored as null, so
		// removing a field leaves nothing behind.
		if value == nil {
			continue
		}

		if field.Unique {
			taken, err := st.FieldValueTaken(ctx, field.Name, value, self)
			if err != nil {
				return nil, err
			}

			if taken {
				return nil, respond.Fault{
					Status:  http.StatusConflict,
					Message: field.Name + ": another user already has that value",
				}
			}
		}

		data[field.Name] = value
	}

	return data, nil
}
