package schema

import (
	"context"
	"time"

	ev "gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/ent/hook"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/go-kratos/kratos/v2/log"
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
	}
}

// Edges of the OneTimePassword.
func (OneTimePassword) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}

func (OneTimePassword) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(mutator ent.Mutator) ent.Mutator {
				return hook.OneTimePasswordFunc(func(ctx context.Context, mutation *ev.OneTimePasswordMutation) (ent.Value, error) {
					// Mutate the one time password
					v, err := mutator.Mutate(ctx, mutation)
					if err != nil {
						return nil, err
					}

					// Get the client from the context
					client := ev.FromContext(ctx)
					if client == nil {
						log.Errorf("[hook:otp] client is not found in context")
						return v, nil
					}

					// if mutation has isUsed = true value
					isUsed, ok := mutation.IsUsed()
					if !ok || !isUsed {
						return v, nil
					}

					// Get the old user id
					userID, ok := mutation.UserID()
					if !ok {
						oldUserID, err2 := mutation.OldUserID(ctx)
						if err2 != nil {
							log.Errorf("[hook:otp] failed to get old user id: %v", err2)
							return v, nil
						}
						userID = oldUserID
					}

					// Delete the invite link if the member is created or updated to waiting
					_, err2 := client.User.UpdateOneID(userID).
						ClearRemoveAt().
						Save(ctx)
					if err2 != nil {
						log.Errorf("[hook:otp] failed to update user: %v", err2)
					}

					return v, nil
				})
			},
			ent.OpUpdate|ent.OpUpdateOne,
		),
	}
}
