package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramBot struct {
	logger logger.Logger
	name   string
	cfg    *config.TelegramBotConfig
	botApi *tgbotapi.BotAPI
}

func NewTelegramBot(name string, cfg *config.TelegramBotConfig) (*telegramBot, error) {
	botApi, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &telegramBot{
		botApi: botApi,
		cfg:    cfg,
		name:   name,
		logger: logger.Null,
	}, nil
}

func (o *telegramBot) WithLogger(logger logger.Logger) *telegramBot {
	o.logger = logger
	return o
}

func (t *telegramBot) Inform(result *parser.ParseResult, err error) error {
	var msg string
	if err != nil {
		msg = fmt.Sprintf("Result for: %s\n\nError: %s", t.name, err)
	} else {
		if t.cfg.Pretty {
			var prettyJSON bytes.Buffer
			errPretty := json.Indent(&prettyJSON, []byte(result.ToJson()), "", "\t")
			if errPretty != nil {
				t.logger.Errorw("enable to prettify result", "error", errPretty.Error())
				return errPretty
			}
			msg = fmt.Sprintf("Result for: *%s*\n\n```json\n%s\n```", t.name, prettyJSON.String())
		} else {
			msg = fmt.Sprintf("Result for: *%s*\n\n```json\n%s\n```", t.name, result.ToJson())
		}
	}

	for _, id := range t.cfg.UsersId {
		msgForSend := tgbotapi.NewMessage(id, msg)
		msgForSend.ParseMode = "markdown"
		_, errSend := t.botApi.Send(msgForSend)
		if errSend != nil {
			t.logger.Errorw("unable to send result", "error", errSend.Error(), "userId", fmt.Sprintf("%d", id))
		}
	}

	return nil
}
