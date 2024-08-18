//go:build ignore

package main

import (
	"fmt"

	"./checks"

	uptimegopher "uptime-gopher/uptime-gopher"
)

func Setup() uptimegopher.PluginInfo {
	return uptimegopher.PluginInfo{
		Name: "Uptime Gopher Standard Plugin",
		Checks: []uptimegopher.Check{
			// checks.HttpCheck(),
			checks.DomainCheck(),
			checks.SslCheck(),
		},
	}
}

func Shutdown() {
	fmt.Println("shutdown")
}
