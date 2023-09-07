package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("phone").Unique().Nillable().Optional(),
		field.String("email").Unique().Nillable().Optional(),
		field.String("name").Default(""),
		field.String("bio").Default(""),
		field.String("avatar").Nillable().Optional(),
		field.String("timezone").Default("UTC"),
		field.Bool("is_active").Default(false),
		field.Time("last_login_at").Default(time.Now),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
		field.Time("deleted_at").Nillable().Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
