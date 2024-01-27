package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/utils"
	"strconv"
)

type singleRecord struct {
	Name  string          `json:"name,omitempty"`
	Body  json.RawMessage `json:"body,omitempty"`
	Index *int            `json:"index,omitempty"`
	Error *error          `json:"error,omitempty"`
}

type Notifier interface {
	notify(*singleRecord) error

	Inform(result *parser.ParseResult, err error, asArray bool) error
}

func resultToSingleRecord(name string, result *parser.ParseResult, errResult error, logger logger.Logger) *singleRecord {
	if errResult != nil {
		return &singleRecord{
			Name:  name,
			Error: &errResult,
		}
	}

	body, errMarshal := json.Marshal(result.ToJson())
	if errMarshal != nil {
		logger.Errorw("cant unmarshal body", "error", errMarshal.Error())
		return &singleRecord{
			Name:  name,
			Error: &errMarshal,
		}
	}
	return &singleRecord{
		Name: name,
		Body: body,
	}
}

func resultToSingleArray(name string, result *parser.ParseResult, errResult error, logger logger.Logger) ([]*singleRecord, error) {
	var arr []interface{}

	err := json.Unmarshal([]byte(result.ToJson()), &arr)
	if err != nil {
		logger.Errorw("unable to unmarshal result like array", "error", err.Error())
		return nil, err
	}

	records := make([]*singleRecord, len(arr))

	for i, v := range arr {
		index := i
		lv := v

		body, errMarshal := json.Marshal(lv)
		if errMarshal != nil {
			logger.Errorw("unable to unmarshal like array", "error", errMarshal.Error())
			records[index] = &singleRecord{
				Name:  name,
				Body:  nil,
				Error: &errMarshal,
				Index: &index,
			}
			continue
		}
		records[index] = &singleRecord{
			Name:  name,
			Body:  body,
			Index: &index,
		}
	}

	return records, nil
}

func ShouldInform(cfg *config.NotifierConfig, result builder.Jsonable) (bool, error) {
	if cfg.Force || cfg.Expression == "" {
		return true, nil
	}

	out, err := utils.ProcessExpression(cfg.Expression, result, nil)
	if err != nil {
		return false, err
	}

	value, err := strconv.ParseBool(fmt.Sprintf("%v", out))
	if err != nil {
		return false, err
	}

	return value, nil
}

func inform(notifier Notifier, name string, result *parser.ParseResult, errResult error, asArray bool, logger logger.Logger) error {
	if errResult != nil {
		return notifier.notify(resultToSingleRecord(name, nil, errResult, logger))
	}
	if !asArray {
		return notifier.notify(resultToSingleRecord(name, result, nil, logger))
	}

	records, err := resultToSingleArray(name, result, errResult, logger)
	if err != nil {
		return err
	}

	for _, rec := range records {
		errNotify := notifier.notify(rec)
		if errNotify != nil {
			return errNotify
		}
	}

	return nil
}
