package logger

import "context"

type requestIDKeyType string

const requestIDKey requestIDKeyType = "request-id"

// NewContext returns a new Context that carries value u.
func NewContext(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, requestIDKey, reqID)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) string {
	v, ok := ctx.Value(requestIDKey).(string)
	if ok {
		return v
	}
	return ""
}

// RequestIDMiddleware logger middleware that adds the request id as log field
// if present in the context.
func RequestIDMiddleware(ctx context.Context) []interface{} {
	if reqID := FromContext(ctx); reqID != "" {
		return []interface{}{"request_id", reqID}
	}
	return nil
}
