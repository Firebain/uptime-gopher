package main

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"reflect"
	"strconv"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type DynamicPlugin struct {
	id string

	yaegi *interp.Interpreter

	name       string
	setupFn    func(*PluginCtx) error
	shutdownFn func(*PluginCtx) error
}

func (d *DynamicPlugin) Id() string {
	return d.id
}

func (d *DynamicPlugin) Name() string {
	return d.name
}

func (d *DynamicPlugin) Setup(app *PluginCtx) error {
	return d.setupFn(app)
}

func (d *DynamicPlugin) Shutdown(app *PluginCtx) error {
	return d.shutdownFn(app)
}

func NewDynamicPlugin(fs fs.FS) (*DynamicPlugin, error) {
	yaegi := interp.New(interp.Options{
		SourcecodeFilesystem: fs,
		GoPath:               "./_pkg",
	})

	yaegi.Use(stdlib.Symbols)
	yaegi.Use(Symbols)

	_, err := yaegi.EvalPath("plugin.go")
	if err != nil {
		return nil, err
	}

	name, err := yaegi.Eval("Name")
	if err != nil {
		return nil, fmt.Errorf("failed to eval Name: %w", err)
	}

	if name.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("Name must be string")
	}

	nameStr := name.String()

	setup, err := yaegi.Eval("Setup")
	if err != nil {
		return nil, err
	}

	setupFn, ok := setup.Interface().(func(*PluginCtx) error)
	if !ok {
		return nil, fmt.Errorf("Setup() must be `func (*PluginCtx) error`")
	}

	shutdown, err := yaegi.Eval("Shutdown")
	if err != nil {
		return nil, err
	}

	shutdownFn, ok := shutdown.Interface().(func(*PluginCtx) error)
	if !ok {
		return nil, fmt.Errorf("Shutdown() must be `func (*PluginCtx) error`")
	}

	return &DynamicPlugin{
		id: strconv.FormatUint(rand.Uint64(), 10),

		yaegi: yaegi,

		name:       nameStr,
		setupFn:    setupFn,
		shutdownFn: shutdownFn,
	}, nil
}

func LoadDynamicPluginsFromDir(dir string) ([]*DynamicPlugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	plugins := []*DynamicPlugin{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folder := fmt.Sprintf("%s/%s", dir, entry.Name())

		plugin, err := NewDynamicPlugin(os.DirFS(folder))
		if err != nil {
			return nil, err
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}
