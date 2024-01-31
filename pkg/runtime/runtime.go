package runtime

import (
	"context"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/registry"
	"github.com/PxyUp/fitter/pkg/trigger"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
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
	r.runHTTPServer(updates)
	<-r.ctx.Done()
	close(updates)
}

func (r *runtime) createRunTime(updates <-chan string) {
	reg := registry.NewFromConfig(r.cfg, r.logger.With("registry", "runtime"))
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case n := <-updates:
				lName := n
				go func(name string) {
					r.logger.Infow("new trigger comes", "name", name)
					_, _ = reg.Get(name).Process()
				}(lName)
			}
		}
	}()
}

func (r *runtime) runHTTPServer(updates chan<- string) {
	needRun := false
	for _, item := range r.cfg.Items {
		if item.TriggerConfig != nil && item.TriggerConfig.HTTPTrigger != nil {
			needRun = true
			break
		}
	}

	if !needRun {
		return
	}
	if r.cfg.HttpServer == nil || r.cfg.HttpServer.Port == 0 {
		log.Fatalf("port for http server not setup")
		return
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	port := fmt.Sprintf(":%d", r.cfg.HttpServer.Port)
	path := "/trigger/:name"
	engine.POST(path, func(c *gin.Context) {
		n := c.Param("name")
		go func(name string) {
			updates <- name
		}(n)
		c.Status(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    port,
		Handler: engine,
	}
	go func() {
		<-r.ctx.Done()
		r.logger.Debug("starting graceful shutdown for http_server")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		r.logger.Debug("graceful shutdown done for http_server")

	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	r.logger.Infow("start http server...", "port", port)
	r.logger.Infow("you send POST request for trigger", "path", path)
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
