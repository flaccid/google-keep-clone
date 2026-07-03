package store

import "context"

type contextKey string

const (
	ownerKey    contextKey = "keep_owner"
	emailKey    contextKey = "keep_email"
	avatarURLKey contextKey = "keep_avatar_url"
)

func WithOwner(ctx context.Context, owner string) context.Context {
	return context.WithValue(ctx, ownerKey, owner)
}

func OwnerFromContext(ctx context.Context) string {
	owner, _ := ctx.Value(ownerKey).(string)
	return owner
}

func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey, email)
}

func EmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(emailKey).(string)
	return email
}

func WithAvatarURL(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, avatarURLKey, url)
}

func AvatarURLFromContext(ctx context.Context) string {
	url, _ := ctx.Value(avatarURLKey).(string)
	return url
}
