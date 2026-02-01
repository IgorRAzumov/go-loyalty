package authctx

import "context"

type userIDContextKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserID возвращает userID из context, если он установлен.
func UserID(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	value := ctx.Value(userIDContextKey{})
	id, ok := value.(int64)
	return id, ok && id > 0
}
