package middleware

type Chain struct {
	middleware []Middleware
}

func NewChain() *Chain {
	return &Chain{
		middleware: make([]Middleware, 0),
	}
}

func (c *Chain) Use(m Middleware) *Chain {
	c.middleware = append(c.middleware, m)
	return c
}

func (c *Chain) Execute(ctx *Context, final HandlerFunc) error {
	ctx.middleware = c.middleware
	ctx.index = -1

	finalMiddleware := func(ctx *Context, next func() error) error {
		return final(ctx)
	}

	allMiddleware := append(c.middleware, finalMiddleware)
	ctx.middleware = allMiddleware

	return ctx.Next()
}

func Sequence(middlewares ...Middleware) []Middleware {
	return middlewares
}
