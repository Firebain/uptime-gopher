package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"gopkg.in/yaml.v3"
)

type Job struct {
	domain      Domain
	check       Check
	checkConfig CheckConfig
	next        time.Time
}

type PluginCtx struct {
	id  string
	app *App
}

func (p *PluginCtx) AddCheck(check Check) {
	p.app.AddCheck(p.id, check)
}

type Plugin interface {
	Id() string
	Name() string

	Setup(*PluginCtx) error
	Shutdown(*PluginCtx) error
}

type App struct {
	log *slog.Logger

	plugins []Plugin
	checks  map[string]map[string]Check
}

func NewApp(log *slog.Logger) *App {
	log = log.With("service", "App")

	return &App{
		log: log,

		plugins: []Plugin{},
		checks:  map[string]map[string]Check{},
	}
}

func (a *App) AddPlugin(plugin Plugin) error {
	a.log.Info("Loading plugin", "name", plugin.Name())

	ctx := PluginCtx{
		id:  plugin.Id(),
		app: a,
	}

	err := plugin.Setup(&ctx)
	if err != nil {
		return err
	}

	a.plugins = append(a.plugins, plugin)

	a.log.Info("Plugin loaded", "name", plugin.Name())

	return nil
}

func (a *App) AddCheck(namespace string, check Check) {
	checkExists, _ := a.GetCheck(check.Key)
	if checkExists != nil {
		a.log.Warn("Check already exists. Skippting", "name", check.Key, "namespace", namespace)

		return
	}

	namespaceVals, ok := a.checks[namespace]
	if !ok {
		namespaceVals = map[string]Check{}
	}
	namespaceVals[check.Key] = check

	a.checks[namespace] = namespaceVals
}

func (a *App) GetCheck(key string) (*Check, error) {
	for _, namespace := range a.checks {
		if check, ok := namespace[key]; ok {
			return &check, nil
		}
	}

	return nil, fmt.Errorf("check not found")
}

func main() {
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		TimeFormat: time.DateTime,
		Level:      slog.LevelDebug,
	}))

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Error("Failed to read config file", "error", err)

		os.Exit(1)
	}

	log.Info("Loading config...")

	config := Config{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Error("Failed to unmarshal config", "error", err)

		os.Exit(1)
	}

	log.Info("Config loaded")

	app := NewApp(log)

	log.Info("Loading plugins...")

	plugins, err := LoadDynamicPluginsFromDir("./plugins")
	if err != nil {
		log.Error("Failed to load plugins", "error", err)

		os.Exit(1)
	}

	for _, plugin := range plugins {
		err := app.AddPlugin(plugin)
		if err != nil {
			log.Error("Failed to setup plugin", "name", plugin.Name(), "error", err)

			os.Exit(1)
		}
	}

	log.Info("Validate config...")

	for _, domain := range config.Domains {
		for _, checkConfig := range domain.Checks {
			check, _ := app.GetCheck(checkConfig.Key)
			if check == nil {
				log.Error("Check not found", "name", checkConfig.Key, "domain", domain.Domain)

				os.Exit(1)
			}

			if check.ValidateArgs != nil {
				err := check.ValidateArgs(checkConfig.Args)
				if err != nil {
					log.Error("Check args validation failed", "name", check.Name, "domain", domain.Domain, "error", err)

					os.Exit(1)
				}
			}
		}
	}

	log.Info("Config validated")

	jobs := []*Job{}
	for _, domain := range config.Domains {
		for _, checkConfig := range domain.Checks {
			check, err := app.GetCheck(checkConfig.Key)
			if err != nil {
				log.Error("Check not found", "name", checkConfig.Key, "domain", domain.Domain)

				os.Exit(1)
			}

			log.Info("Adding job", "name", check.Name, "domain", domain.Domain, "args", checkConfig.Args)

			jobs = append(jobs, &Job{
				domain:      domain,
				check:       *check,
				checkConfig: checkConfig,
				next:        time.Now(),
			})
		}
	}

	log.Info("Starting scheduler...")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second)

mainloop:
	for {
		select {
		case <-exit:
			log.Info("Exiting...")

			break mainloop
		case <-ticker.C:
			now := time.Now()

			for _, job := range jobs {
				if !job.next.Before(now) {
					continue
				}

				log.Info("Running check", "name", job.check.Name, "domain", job.domain.Domain)

				result := job.check.Run(job.domain.Domain, job.checkConfig.Args)

				if !result.Success {
					if result.Severity == SeverityDebug {
						log.Debug("Check Debug", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)
					}

					if result.Severity == SeverityNotice {
						log.Info("Check Notice", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)
					}

					if result.Severity == SeverityWarning {
						log.Warn("Check Warning", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)

					}

					if result.Severity == SeverityError {
						log.Error("Check Error", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)
					}

					if result.Severity == SeverityDown {
						log.Error("Domain Down", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)
					}

					if result.Severity == SeverityFatal {
						log.Error("Check Fatal", "name", job.check.Name, "domain", job.domain.Domain, "message", result.Message)

						break mainloop
					}
				}

				if job.checkConfig.Interval > 0 {
					job.next = now.Add(job.checkConfig.Interval)

					continue
				}

				if job.domain.Interval > 0 {
					job.next = now.Add(job.domain.Interval)

					continue
				}

				job.next = now.Add(time.Minute)
			}
		}
	}
}
