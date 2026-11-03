package xiao

type keySession struct{}

// SetSession 设置会话名，存于Env，由派生的Context继承
func SetSession(ctx Context, session string) {
	ctx.Set(keySession{}, session)
}

// GetSession 返回会话名，未设置时以Context的名称代替
func GetSession(ctx Context) string {
	if val, ok := ctx.Get(keySession{}); ok {
		if session := val.(string); session != "" {
			return session
		}
	}
	return ctx.Name()
}

// GetRealSession 返回经SetSession设置的会话名，未设置时为空
func GetRealSession(ctx Context) string {
	return ctx.GetString(keySession{})
}
