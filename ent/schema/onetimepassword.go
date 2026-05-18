package schema

import (
	"time"

	"github.com/makesalekz/iam/ent/enum"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OneTimePassword holds the schema definition for the OneTimePassword entity.
type OneTimePassword struct {
	ent.Schema
}

// Fields of the OneTimePassword.
func (OneTimePassword) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("code").MaxLen(6),
		field.Enum("type").GoType(enum.OneTimePasswordType("")),
		field.Bool("is_used").Default(false),
		field.Time("expires_at"),
		field.Time("created_at").Default(time.Now),
		field.Int64("failed_attempts").Default(0),
	}
}

// Edges of the OneTimePassword.
func (OneTimePassword) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}
