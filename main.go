package main

import (
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

// Uptime Gopher

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

	log.Info("Loading plugins...")

	plugins, err := LoadPluginsFromDir("./plugins")
	if err != nil {
		log.Error("Failed to load plugins", "error", err)

		os.Exit(1)
	}

	checks := map[string]Check{}
	for _, plugin := range plugins {
		log.Info("Plugin loaded", "name", plugin.Info.Name)

		for _, check := range plugin.Info.Checks {
			checks[check.Key] = check

			log.Info("Check loaded", "name", check.Name, "key", check.Key)
		}
	}

	log.Info("Validate config...")

	for _, domain := range config.Domains {
		for _, checkConfig := range domain.Checks {
			check, ok := checks[checkConfig.Key]
			if !ok {
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
			check := checks[checkConfig.Key]

			log.Info("Adding job", "name", check.Name, "domain", domain.Domain, "args", checkConfig.Args)

			jobs = append(jobs, &Job{
				domain:      domain,
				check:       checks[check.Key],
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
