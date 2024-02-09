package notifier

import (
	"context"
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	_ Notifier = &redisNotifier{}
)

type redisNotifier struct {
	logger logger.Logger
	name   string
	cfg    *config.RedisNotifierConfig
}

func (r *redisNotifier) notify(record *singleRecord, input builder.Interfacable) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     formatWithRecord(r.cfg.Addr, record, input),
		Password: formatWithRecord(r.cfg.Password, record, input),
		DB:       r.cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	msg, errMars := json.Marshal(record)
	if errMars != nil {
		r.logger.Errorw("cant marshal message", "error", errMars.Error())
		return errMars
	}

	errSend := redisClient.Publish(ctx, formatWithRecord(r.cfg.Channel, record, input), msg).Err()
	if errSend != nil {
		r.logger.Errorw("cant send message", "error", errSend.Error())
		return errSend
	}

	return redisClient.Close()
}

func (o *redisNotifier) GetLogger() logger.Logger {
	return o.logger
}

func (r *redisNotifier) WithLogger(logger logger.Logger) *redisNotifier {
	r.logger = logger
	return r
}

func NewRedis(name string, cfg *config.RedisNotifierConfig) *redisNotifier {
	return &redisNotifier{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}
