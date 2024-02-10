package trigger

import (
	"context"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

type Message struct {
	Name  string
	Value builder.Interfacable
}

type Trigger interface {
	Run(chan<- *Message)
	Stop()
}

func createHttpTrigger(ctx context.Context, cfg *config.Config, logger logger.Logger) []Trigger {
	needRun := false
	forIgnore := []string{}
	for _, item := range cfg.Items {
		if item.TriggerConfig != nil && item.TriggerConfig.HTTPTrigger != nil {
			needRun = true
		} else {
			forIgnore = append(forIgnore, item.Name)
		}
	}

	if !needRun {
		return nil
	}

	return []Trigger{HttpServer(ctx, cfg.HttpServer, forIgnore).WithLogger(logger.With("scheduler_type", "http_server"))}
}

func createSchedulerTriggers(ctx context.Context, cfg *config.Config, logger logger.Logger) []Trigger {
	var schedulers []Trigger
	for _, item := range cfg.Items {
		if item.TriggerConfig != nil && item.TriggerConfig.SchedulerTrigger != nil {
			schedulers = append(schedulers, Scheduler(ctx, item.Name, item.TriggerConfig.SchedulerTrigger).WithLogger(logger.With("scheduler_name", item.Name)))
		}
	}

	return schedulers
}

func CreateTriggers(ctx context.Context, cfg *config.Config, logger logger.Logger) []Trigger {
	var triggers []Trigger
	triggers = append(triggers, createHttpTrigger(ctx, cfg, logger)...)
	triggers = append(triggers, createSchedulerTriggers(ctx, cfg, logger)...)

	return triggers
}
