//go:build ignore

package main

import (
	"fmt"

	"./checks"

	uptimegopher "uptime-gopher/uptime-gopher"
)

var Name = "Uptime Gopher Standard Plugin"

func Setup(ctx *uptimegopher.PluginCtx) error {
	// ctx.AddCheck(checks.HttpCheck())
	ctx.AddCheck(checks.DomainCheck())
	ctx.AddCheck(checks.SslCheck())

	return nil
}

func Shutdown(ctx *uptimegopher.PluginCtx) error {
	fmt.Println("shutdown")
}
