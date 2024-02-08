package notifier

import (
	"context"
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	_ Notifier = &redisNotifier{}
)

type redisNotifier struct {
	logger      logger.Logger
	name        string
	cfg         *config.RedisNotifierConfig
	redisClient *redis.Client
}

func (r *redisNotifier) notify(record *singleRecord, input builder.Interfacable) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	msg, errMars := json.Marshal(record)
	if errMars != nil {
		r.logger.Errorw("cant marshal message", "error", errMars.Error())
		return errMars
	}

	var channel string
	if record.Error != nil {
		channel = utils.Format(r.cfg.Channel, builder.String((*record.Error).Error()), record.Index, input)
	} else {
		channel = utils.Format(r.cfg.Channel, builder.ToJsonable(record.Body), record.Index, input)
	}

	errSend := r.redisClient.Publish(ctx, channel, msg).Err()
	if errSend != nil {
		r.logger.Errorw("cant send message", "error", errSend.Error())
		return errSend
	}
	return nil
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
		redisClient: redis.NewClient(&redis.Options{
			Addr:     utils.Format(cfg.Addr, nil, nil, nil),
			Password: utils.Format(cfg.Password, nil, nil, nil),
			DB:       cfg.DB,
		}),
	}
}
