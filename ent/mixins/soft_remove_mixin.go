package mixins

import (
	"context"

	"github.com/makesalekz/iam/ent/intercept"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// SoftRemoveMixin implements the soft delete pattern for schemas.
type SoftRemoveMixin struct {
	mixin.Schema
}

// Fields of the SoftRemoveMixin.
func (SoftRemoveMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("remove_at").Nillable().Optional(),
	}
}

type softRemoveKey struct{}

// SkipSoftRemove returns a new context that skips the soft-remove interceptor/mutators.
func SkipSoftRemove(parent context.Context) context.Context {
	return context.WithValue(parent, softRemoveKey{}, true)
}

// Interceptors of the SoftRemoveMixin.
func (d SoftRemoveMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(softRemoveKey{}).(bool); skip {
				return nil
			}
			d.P(q)
			return nil
		}),
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d SoftRemoveMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}
