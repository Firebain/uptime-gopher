package main

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Severity int

const (
	SeverityDebug Severity = iota
	SeverityNotice
	SeverityWarning
	SeverityError
	SeverityDown
	SeverityFatal
)

type CheckResult struct {
	Success  bool
	Severity Severity
	Message  string
}

type Check struct {
	Key          string
	Name         string
	Run          func(string, map[string]string) CheckResult
	ValidateArgs func(map[string]string) error
}

type PluginInfo struct {
	Key    string
	Name   string
	Checks []Check
}

type Plugin struct {
	yaegi *interp.Interpreter
	Info  PluginInfo
}

func NewPluginFromFs(fs fs.FS) (*Plugin, error) {
	yaegi := interp.New(interp.Options{
		SourcecodeFilesystem: fs,
		GoPath:               "./_pkg",
	})

	yaegi.Use(stdlib.Symbols)
	yaegi.Use(Symbols)
	yaegi.Use(map[string]map[string]reflect.Value{
		"uptime-gopher/uptime-gopher": {
			"PluginInfo":  reflect.ValueOf((*PluginInfo)(nil)),
			"Check":       reflect.ValueOf((*Check)(nil)),
			"CheckResult": reflect.ValueOf((*CheckResult)(nil)),

			"Severity":        reflect.ValueOf((*Severity)(nil)),
			"SeverityDebug":   reflect.ValueOf(SeverityDebug),
			"SeverityNotice":  reflect.ValueOf(SeverityNotice),
			"SeverityWarning": reflect.ValueOf(SeverityWarning),
			"SeverityError":   reflect.ValueOf(SeverityError),
			"SeverityDown":    reflect.ValueOf(SeverityDown),
			"SeverityFatal":   reflect.ValueOf(SeverityFatal),
		},
	})

	_, err := yaegi.EvalPath("plugin.go")
	if err != nil {
		return nil, err
	}

	res, err := yaegi.Eval("Setup()")
	if err != nil {
		return nil, err
	}

	info, ok := res.Interface().(PluginInfo)
	if !ok {
		return nil, fmt.Errorf("Setup() must return PluginInfo")
	}

	return &Plugin{
		yaegi: yaegi,
		Info:  info,
	}, nil
}

func LoadPluginsFromDir(dir string) ([]*Plugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	plugins := []*Plugin{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folder := fmt.Sprintf("%s/%s", dir, entry.Name())

		plugin, err := NewPluginFromFs(os.DirFS(folder))
		if err != nil {
			return nil, err
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}
