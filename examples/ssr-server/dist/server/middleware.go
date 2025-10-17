package main

import (
	"time"

	"github.com/galaxy/galaxy/pkg/middleware"
)

func OnRequest(ctx *middleware.Context, next func() error) error {
	ctx.Set("timestamp", time.Now().Format(time.RFC3339))
	ctx.Set("serverName", "Galaxy SSR")
	return next()
}
