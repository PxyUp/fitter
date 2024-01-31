package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramBot struct {
	logger logger.Logger
	name   string
	cfg    *config.TelegramBotConfig
	botApi *tgbotapi.BotAPI
}

func (t *telegramBot) notify(record *singleRecord) error {
	msg, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if t.cfg.Pretty {
		return t.sendMessage(t.prettyMsg(string(msg)))
	}

	return t.sendMessage(string(msg))
}

var (
	_ Notifier = &telegramBot{}
)

func NewTelegramBot(name string, cfg *config.TelegramBotConfig) (*telegramBot, error) {
	botApi, err := tgbotapi.NewBotAPI(utils.Format(cfg.Token, nil, nil))
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

func (t *telegramBot) sendMessage(msg string) error {
	for _, id := range t.cfg.UsersId {
		msgForSend := tgbotapi.NewMessage(id, msg)
		_, errSend := t.botApi.Send(msgForSend)
		if errSend != nil {
			t.logger.Errorw("unable to send result", "error", errSend.Error(), "userId", fmt.Sprintf("%d", id))
		}
	}

	return nil
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

func (t *telegramBot) Inform(result *parser.ParseResult, errResult error, asArray bool) error {
	return inform(t, t.name, result, errResult, asArray, t.logger)
}
