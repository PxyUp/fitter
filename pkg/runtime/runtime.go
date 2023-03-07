package runtime

import (
	"context"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/registry"
	"github.com/PxyUp/fitter/pkg/trigger"
)

type runtime struct {
	ctx    context.Context
	cfg    *config.Config
	logger logger.Logger
}

func New(ctx context.Context, cfg *config.Config, logger logger.Logger) *runtime {
	return &runtime{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}
}

func (r *runtime) Start() {
	updates := make(chan string)
	r.createRunTime(updates)
	r.runScheduler(updates)
	<-r.ctx.Done()
}

func (r *runtime) createRunTime(updates <-chan string) {
	reg := registry.NewFromConfig(r.cfg, r.logger.With("registry", "runtime"))
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case name := <-updates:
				_, _ = reg.Get(name).Process()
			}
		}
	}()
}

func (r *runtime) runScheduler(updates chan<- string) {
	for _, item := range r.cfg.Items {
		if item.TriggerConfig != nil && item.TriggerConfig.SchedulerTrigger != nil {
			localTrigger := trigger.Scheduler(r.ctx, item.Name, item.TriggerConfig.SchedulerTrigger).WithLogger(r.logger.With("scheduler_name", item.Name))
			localTrigger.Run(updates)
			go func() {
				<-r.ctx.Done()
				localTrigger.Stop()
			}()
		}
	}
}
