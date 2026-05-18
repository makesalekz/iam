package schema

import (
	"github.com/makesalekz/iam/ent/mixins"
	u_struc "github.com/makesalekz/utils/v2/struc"

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
		field.Int64("external_user_id").Optional().Nillable(),
		field.String("mail").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("display_name").Optional().Nillable(),
		field.Enum("provider").GoType(u_struc.Provider("")).Optional().Nillable(),
		field.String("access_token"),
		field.String("token_type").Optional().Nillable(),
		field.String("refresh_token").Optional().Nillable(),
		field.Time("expires_at").Optional().Nillable(),
	}
}

func (UserCredentials) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}

func (UserCredentials) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
		mixins.SoftDeleteMixin{},
	}
}
