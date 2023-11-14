package schema

import (
	"time"

	"gitlab.calendaria.team/services/iam/ent/property"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserSettings holds the schema definition for the UserSettings entity.
type UserSettings struct {
	ent.Schema
}

// Fields of the UserSettings.
func (UserSettings) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Enum("setting").GoType(property.Settings("")),
		field.String("value"),
		field.Time("updated_at").Default(time.Now),
	}
}

// Edges of the UserSettings.
func (UserSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}

// Indexes of the UserPrivacy.
func (UserSettings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "setting").Unique(),
	}
}
