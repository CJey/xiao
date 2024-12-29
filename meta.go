package xiao

type keySession struct{}

// SetSession use to set real session name
func SetSession(ctx Context, session string) {
	ctx.Set(keySession{}, session)
}

// GetSession use to get session name if it has real session, otherwise return context name instead
func GetSession(ctx Context) string {
	if val, ok := ctx.Get(keySession{}); ok {
		if session := val.(string); session != "" {
			return session
		}
	}
	return ctx.Name()
}

// GetRealSession use to get real session name
func GetRealSession(ctx Context) string {
	return ctx.GetString(keySession{})
}
