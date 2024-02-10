package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type httpServer struct {
	ctx    context.Context
	cancel context.CancelFunc

	serverCfg     *config.HttpServerCfg
	logger        logger.Logger
	ignoreTrigger []string
}

func (s *httpServer) WithLogger(logger logger.Logger) *httpServer {
	s.logger = logger
	return s
}

func HttpServer(parentCtx context.Context, serverCfg *config.HttpServerCfg, ignoreTrigger []string) *httpServer {
	ctx, cancel := context.WithCancel(parentCtx)
	return &httpServer{
		serverCfg:     serverCfg,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger.Null,
		ignoreTrigger: ignoreTrigger,
	}
}

func (s *httpServer) Run(updates chan<- *Message) {
	if s.serverCfg == nil || s.serverCfg.Port == 0 {
		log.Fatalf("port for http server not setup")
		return
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	port := fmt.Sprintf(":%d", s.serverCfg.Port)
	path := "/trigger/:name"
	engine.POST(path, func(c *gin.Context) {
		msg := json.RawMessage{}

		errBind := c.BindJSON(&msg)
		if errBind != nil {
			s.logger.Errorw("cant bind request data", "error", errBind.Error())
			c.Status(http.StatusBadRequest)
			return
		}

		n := c.Param("name")
		go func(name string, value json.RawMessage) {
			for _, v := range s.ignoreTrigger {
				if v == name {
					s.logger.Debugw("ignoring trigger", "name", v)
					return
				}
			}

			updates <- &Message{
				Name:  name,
				Value: builder.ToJsonable(value),
			}
		}(n, msg)
		c.Status(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    port,
		Handler: engine,
	}
	go func() {
		<-s.ctx.Done()
		s.logger.Debug("starting graceful shutdown for http_server")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		s.logger.Debug("graceful shutdown done for http_server")

	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	s.logger.Infow("start http server...", "port", port)
	s.logger.Infow("now you send POST request for trigger", "path", path)
}

func (s *httpServer) Stop() {
	if s.ctx == nil {
		return
	}

	s.cancel()
	s.ctx = nil
	s.cancel = nil
}
