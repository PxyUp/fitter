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
	triggers := trigger.CreateTriggers(r.ctx, r.cfg, r.logger)
	r.createRunTime(updates)
	for _, t := range triggers {
		t.Run(updates)
	}
	<-r.ctx.Done()
	close(updates)
	for _, t := range triggers {
		t.Stop()
	}
}

func (r *runtime) createRunTime(updates <-chan *trigger.Message) {
	reg := registry.NewFromConfig(r.cfg, r.logger.With("registry", "runtime"))
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case n, ok := <-updates:
				if !ok {
					return
				}
				lName := n
				go func(name string, value builder.Interfacable) {
					fields := []string{"name", name}
					if value != nil {
						fields = append(fields, "input", value.ToJson())
					}
					r.logger.Infow("new trigger comes", fields...)
					_, _ = reg.Get(name).Process(value)
				}(lName.Name, lName.Value)
			}
		}
	}()
}
