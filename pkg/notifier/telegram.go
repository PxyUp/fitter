package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramBot struct {
	logger logger.Logger
	name   string
	cfg    *config.TelegramBotConfig
}

func (t *telegramBot) notify(record *singleRecord, input builder.Interfacable) error {
	var msg []byte
	value := recordToInterfacable(record)
	if t.cfg.OnlyMsg {
		if record.Error != nil {
			msg = value.Raw()
		} else {
			msg = record.Body
		}
	} else {
		body, err := json.Marshal(record)
		if err != nil {
			return err
		}

		msg = body
	}

	if t.cfg.Pretty {
		return t.sendMessage(t.prettyMsg(string(msg)), record, input)
	}

	return t.sendMessage(string(msg), record, input)
}

var (
	_ Notifier = &telegramBot{}
)

func NewTelegramBot(name string, cfg *config.TelegramBotConfig) *telegramBot {
	return &telegramBot{
		cfg:    cfg,
		name:   name,
		logger: logger.Null,
	}
}

func (o *telegramBot) WithLogger(logger logger.Logger) *telegramBot {
	o.logger = logger
	return o
}

func (t *telegramBot) sendMessage(msg string, record *singleRecord, input builder.Interfacable) error {
	botApi, err := tgbotapi.NewBotAPI(formatWithRecord(t.cfg.Token, record, input))
	if err != nil {
		return err
	}
	botApi.Client = http_client.GetDefaultClient()

	for _, id := range t.cfg.UsersId {
		msgForSend := tgbotapi.NewMessage(id, msg)
		_, errSend := botApi.Send(msgForSend)
		if errSend != nil {
			t.logger.Errorw("unable to send result", "error", errSend.Error(), "userId", fmt.Sprintf("%d", id))
		}
	}

	return nil
}

func (o *telegramBot) GetLogger() logger.Logger {
	return o.logger
}

func (t *telegramBot) prettyMsg(msg string) string {
	if !t.cfg.Pretty {
		return msg
	}
	var prettyJSON bytes.Buffer
	errPretty := json.Indent(&prettyJSON, []byte(msg), "", " ")
	if errPretty != nil {
		t.logger.Errorw("unable to prettify result", "error", errPretty.Error())
		return msg
	}
	return prettyJSON.String()
}
