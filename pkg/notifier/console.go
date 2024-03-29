package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"os"
)

type console struct {
	logger logger.Logger
	name   string
	cfg    *config.ConsoleConfig
}

func (o *console) GetLogger() logger.Logger {
	return o.logger
}

func (o *console) notify(record *singleRecord, input builder.Interfacable) error {
	if o.cfg.OnlyResult {
		_, errOut := fmt.Fprintln(os.Stdout, string(record.Body))
		if errOut != nil {
			o.logger.Errorw("cant send to stdout", "error", errOut.Error())
			return errOut
		}
		return nil
	}

	bb, err := json.Marshal(record)
	if err != nil {
		o.logger.Errorw("cant unmarshal record", "error", err.Error())
		return err
	}

	_, errOut := fmt.Fprintln(os.Stdout, string(bb))
	if errOut != nil {
		o.logger.Errorw("cant send to stdout", "error", errOut.Error())
		return errOut
	}

	return nil
}

var (
	_ Notifier = &console{}
)

func NewConsole(name string, cfg *config.ConsoleConfig) *console {
	return &console{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}

func (o *console) WithLogger(logger logger.Logger) *console {
	o.logger = logger
	return o
}
