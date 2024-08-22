package schema

import (
	"time"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/index"

	"gitlab.calendaria.team/services/iam/ent/mixins"

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
		field.Int64("id").Immutable(),
		field.String("phone").Unique().Nillable().Optional().Comment("phone of a user"),
		field.String("email").Unique().Nillable().Optional().Comment("email of a user"),
		field.String("username").MinLen(3).MaxLen(25).Unique().Nillable().Optional().Comment("username of a user"),
		field.String("name").Default("").Comment("this field contains a name that user set up"),
		field.String("bio").Default("").Comment("this field a biography of a user"),
		field.String("avatar").Nillable().Optional().Comment("a string contains link to a profile pic"),
		field.String("timezone").Default("UTC").Comment("the timezone of a user"),
		field.Bool("is_active").Default(false).Comment("this field indicates that user finished his signup"),
		field.Bool("phone_verified").Default(false).Comment("this field indicates that phone has been verified"),
		field.Bool("email_verified").Default(false).Comment("this field indicates that email has been verified"),
		field.Time("last_login_at").Default(time.Now).Comment("the time when user was last logged in"),
		field.Time("created_at").Default(time.Now).Comment("the time when user has been created"),
		field.Time("updated_at").Default(time.Now).Comment("the time when user was last changed"),
		field.Time("bio_updated_at").Nillable().Optional().Comment("the time when user's bio has been changed"),
		field.Int64("default_tenant_id").Nillable().Optional().Comment("default tenant id of user"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.SoftDeleteMixin{},
		mixins.SoftRemoveMixin{},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique().Annotations(entsql.IndexWhere("username is not null")),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
