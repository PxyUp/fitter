package runtime

import (
	"context"
	"github.com/PxyUp/fitter/pkg/builder"
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
	updates := make(chan *trigger.Message)
	r.createRunTime(updates)
	r.runScheduler(updates)
	r.runHTTPServer(updates)
	<-r.ctx.Done()
	close(updates)
}

func (r *runtime) createRunTime(updates <-chan *trigger.Message) {
	reg := registry.NewFromConfig(r.cfg, r.logger.With("registry", "runtime"))
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case n := <-updates:
				lName := n
				go func(name string, value builder.Jsonable) {
					r.logger.Infow("new trigger comes", "name", name)
					_, _ = reg.Get(name).Process(value)
				}(lName.Name, lName.Value)
			}
		}
	}()
}

func (r *runtime) runHTTPServer(updates chan<- *trigger.Message) {
	needRun := false
	forIgnore := []string{}
	for _, item := range r.cfg.Items {
		if item.TriggerConfig != nil && item.TriggerConfig.HTTPTrigger != nil {
			needRun = true
		} else {
			forIgnore = append(forIgnore, item.Name)
		}
	}

	if !needRun {
		return
	}

	serverRun := trigger.HttpServer(r.ctx, r.cfg.HttpServer, forIgnore).WithLogger(r.logger.With("scheduler_type", "http_server"))
	serverRun.Run(updates)
	go func() {
		<-r.ctx.Done()
		serverRun.Stop()
	}()

}

func (r *runtime) runScheduler(updates chan<- *trigger.Message) {
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
