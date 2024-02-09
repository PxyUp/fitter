package notifier

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/utils"
)

type singleRecord struct {
	Name  string          `json:"name,omitempty"`
	Body  json.RawMessage `json:"body,omitempty"`
	Index *uint32         `json:"index,omitempty"`
	Error *error          `json:"error,omitempty"`
}

type Notifier interface {
	notify(*singleRecord, builder.Interfacable) error
	GetLogger() logger.Logger
}

func recordToInterfacable(record *singleRecord) builder.Interfacable {
	if record.Error != nil {
		return builder.String((*record.Error).Error())
	}
	return builder.ToJsonable(record.Body)
}

func formatWithRecord(text string, record *singleRecord, input builder.Interfacable) string {
	return utils.Format(text, recordToInterfacable(record), record.Index, input)
}

func resultToSingleRecord(name string, result *parser.ParseResult, errResult error, logger logger.Logger) *singleRecord {
	if errResult != nil {
		return &singleRecord{
			Name:  name,
			Error: &errResult,
		}
	}

	return &singleRecord{
		Name: name,
		Body: result.Raw(),
	}
}

func resultToSingleArray(name string, result *parser.ParseResult, errResult error, logger logger.Logger) ([]*singleRecord, error) {
	var arr []interface{}
	err := json.Unmarshal(result.Raw(), &arr)
	if err != nil {
		logger.Errorw("unable to unmarshal result like array", "error", err.Error())
		return nil, err
	}

	records := make([]*singleRecord, len(arr))

	for i, v := range arr {
		index := uint32(i)
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

func ShouldInform(cfg *config.NotifierConfig, result builder.Interfacable) (bool, error) {
	if cfg.Force || cfg.Expression == "" {
		return true, nil
	}

	out, err := utils.ProcessExpression(cfg.Expression, result, nil, nil)
	if err != nil {
		return false, err
	}

	return out.ToInterface() == true, nil
}

func Inform(notifier Notifier, name string, result *parser.ParseResult, errResult error, asArray bool, logger logger.Logger, input builder.Interfacable) error {
	if errResult != nil {
		return notifier.notify(resultToSingleRecord(name, nil, errResult, logger), input)
	}
	if !asArray {
		return notifier.notify(resultToSingleRecord(name, result, nil, logger), input)
	}

	records, err := resultToSingleArray(name, result, errResult, logger)
	if err != nil {
		return err
	}

	for _, rec := range records {
		errNotify := notifier.notify(rec, input)
		if errNotify != nil {
			return errNotify
		}
	}

	return nil
}
