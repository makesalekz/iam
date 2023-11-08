package schema

import (
	"time"

	"gitlab.calendaria.team/alageum-cloud/iam/ent/property"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserPrivacy holds the schema definition for the UserPrivacy entity.
type UserPrivacy struct {
	ent.Schema
}

// Fields of the UserPrivacy.
func (UserPrivacy) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Enum("setting").GoType(property.PrivacySettings("")),
		field.Enum("option").GoType(property.PrivacyOptions("")),
		field.Time("updated_at").Default(time.Now),
	}
}

// Edges of the UserPrivacy.
func (UserPrivacy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}

// Indexes of the UserPrivacy.
func (UserPrivacy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "setting").Unique(),
	}
}
