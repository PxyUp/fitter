package trigger

import (
	"context"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"time"
)

type scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	parentCtx context.Context

	cfg    *config.SchedulerTrigger
	logger logger.Logger
	name   string
}

func Scheduler(parentCtx context.Context, name string, cfg *config.SchedulerTrigger) *scheduler {
	return &scheduler{
		name:      name,
		cfg:       cfg,
		parentCtx: parentCtx,
		logger:    logger.Null,
	}
}

func (s *scheduler) WithLogger(logger logger.Logger) *scheduler {
	s.logger = logger
	return s
}

func (s *scheduler) Run(updates chan<- *Message) {
	if s.ctx != nil {
		return
	}
	localCtx, cancelFn := context.WithCancel(s.parentCtx)

	s.ctx = localCtx
	s.cancel = cancelFn

	go func() {
		if s.cfg.Interval <= 0 {
			s.logger.Info("invalid interval")
			return
		}

		startTime := time.Now()

		updates <- &Message{
			Name:  s.name,
			Value: builder.Int(int(time.Now().Sub(startTime).Seconds())),
		}

		for {
			select {
			case <-localCtx.Done():
				s.logger.Infof("stop scheduler trigger %s", s.name)
				return
			case val := <-time.After(time.Duration(s.cfg.Interval) * time.Second):
				updates <- &Message{
					Name:  s.name,
					Value: builder.Int(int(val.Sub(startTime).Seconds())),
				}
				s.logger.Infof("send scheduled trigger for %s", s.name)
			}
		}
	}()
}

func (s *scheduler) Stop() {
	if s.ctx == nil {
		return
	}

	s.cancel()
	s.ctx = nil
	s.cancel = nil
}
