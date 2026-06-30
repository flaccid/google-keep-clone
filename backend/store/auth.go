package store

import "context"

type contextKey string

const ownerKey contextKey = "keep_owner"

func WithOwner(ctx context.Context, owner string) context.Context {
	return context.WithValue(ctx, ownerKey, owner)
}

func OwnerFromContext(ctx context.Context) string {
	owner, _ := ctx.Value(ownerKey).(string)
	return owner
}
