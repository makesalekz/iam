package schema

import (
	"gitlab.calendaria.team/services/iam/ent/mixins"
	"time"

	"gitlab.calendaria.team/services/iam/ent/enum"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserCredentials holds the schema definition for the UserCredentials entity.
type UserCredentials struct {
	ent.Schema
}

func (UserCredentials) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("mail").Optional().Nillable(),
		field.String("display_name").Optional().Nillable(),
		field.Enum("provider").GoType(enum.Provider("")).Optional().Nillable(),
		field.String("access_token"),
		field.String("token_type").Optional().Nillable(),
		field.String("refresh_token").Optional().Nillable(),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
	}
}

func (UserCredentials) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}

func (UserCredentials) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.SoftDeleteMixin{},
	}
}
