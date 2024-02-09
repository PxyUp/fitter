package notifier

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"os"
)

var (
	_ Notifier = &fileNotifier{}
)

type fileNotifier struct {
	logger logger.Logger
	name   string
	cfg    *config.FileStorageField
}

func (f *fileNotifier) notify(record *singleRecord, input builder.Interfacable) error {
	destinationFileName := formatWithRecord(f.cfg.FileName, record, input)
	destinationPath := formatWithRecord(f.cfg.Path, record, input)

	if f.cfg.Content == "" && len(f.cfg.Raw) == 0 {
		bb, err := json.Marshal(record)
		if err != nil {
			f.logger.Errorw("cannot unmarshal result", "error", err.Error())
			return err
		}

		_, err = utils.CreateFileWithContent(bb, destinationFileName, destinationPath, os.ModePerm, f.cfg.Append, f.logger)
		if err != nil {
			f.logger.Errorw("cannot save result to file", "path", destinationPath, "file_name", destinationFileName, "error", err.Error())
			return err
		}

		return nil
	}
	content := f.cfg.Content
	if len(f.cfg.Raw) > 0 {
		content = string(f.cfg.Raw)
	}

	content = formatWithRecord(content, record, input)
	_, err := utils.CreateFileWithContent([]byte(content), destinationFileName, destinationPath, os.ModePerm, f.cfg.Append, f.logger)
	if err != nil {
		f.logger.Errorw("cannot save result to file", "path", destinationPath, "file_name", destinationFileName, "error", err.Error())
		return err
	}

	return nil
}

func (o *fileNotifier) GetLogger() logger.Logger {
	return o.logger
}

func NewFile(name string, cfg *config.FileStorageField) *fileNotifier {
	return &fileNotifier{
		name:   name,
		cfg:    cfg,
		logger: logger.Null,
	}
}

func (f *fileNotifier) WithLogger(logger logger.Logger) *fileNotifier {
	f.logger = logger
	return f
}
