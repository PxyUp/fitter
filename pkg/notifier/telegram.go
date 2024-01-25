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

var (
	_ Notifier = &telegramBot{}
)

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

func (t *telegramBot) sendError(msg string) error {
	return t.sendMessage(msg)
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

func (t *telegramBot) sendSuccess(result *parser.ParseResult, isArray bool) error {
	if !t.cfg.SendArrayByItem || !isArray {
		return t.sendMessage(fmt.Sprintf("Result for: %s\n\n\n%s", t.name, t.prettyMsg(result.ToJson())))
	}

	var arr []interface{}
	err := json.Unmarshal([]byte(result.ToJson()), &arr)
	if err != nil {
		t.logger.Errorw("unable to unmarshal result like array", "error", err.Error())
		return err
	}

	for _, value := range arr {
		body, errMarshal := json.Marshal(value)
		if errMarshal != nil {
			t.logger.Errorw("unable to unmarshal like array", "error", err.Error())
			continue
		}
		_ = t.sendMessage(fmt.Sprintf("Result for: %s\n\n%s", t.name, t.prettyMsg(string(body))))
	}

	return nil
}

func (t *telegramBot) Inform(result *parser.ParseResult, errResult error, isArray bool) error {
	if errResult != nil {
		return t.sendError(fmt.Sprintf("Result for: %s\n\nError: %s", t.name, errResult))
	}

	return t.sendSuccess(result, isArray)
}
