package parser

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"mime"
	"net/url"
	"os"
	"path/filepath"
)

func filenameFromUrl(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", err
	}
	x, err := url.QueryUnescape(u.EscapedPath())
	if err != nil {
		return "", err
	}
	return filepath.Base(x), nil
}

func CreateFileStorageField(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable, cfg *config.FileStorageField, logger logger.Logger) (string, error) {
	content := cfg.Content
	if len(cfg.Raw) > 0 {
		content = string(cfg.Raw)
	}

	content = utils.Format(content, parsedValue, index, input)
	destinationFileName := utils.Format(cfg.FileName, parsedValue, index, input)
	destinationPath := utils.Format(cfg.Path, parsedValue, index, input)

	return utils.CreateFileWithContent([]byte(content), destinationFileName, destinationPath, os.ModePerm, cfg.Append, logger)
}

func ProcessFileField(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable, field *config.FileFieldConfig, logger logger.Logger) (string, error) {
	destinationFileName := utils.Format(field.FileName, parsedValue, index, input)
	destinationPath := utils.Format(field.Path, parsedValue, index, input)
	destinationURL := utils.Format(field.Url, parsedValue, index, input)

	connector := connectors.NewAPI(destinationURL, field.Config, http_client.GetDefaultClient()).WithLogger(logger.With("connector", "file"))

	headers, body, err := connector.GetWithHeaders(parsedValue, index, input)
	if err != nil {
		logger.Errorw("unable to get file from url", "url", destinationURL, "error", err.Error())
		return "", err
	}

	if destinationFileName == "" {
		_, params, errHeader := mime.ParseMediaType(headers.Get("Content-Disposition"))
		if errHeader != nil {
			filename, errFromUrl := filenameFromUrl(destinationURL)
			if errFromUrl != nil {
				logger.Errorw("unable to set filename for file", "url", destinationURL, "header_value", headers.Get("Content-Disposition"), "error_url", errFromUrl.Error(), "err_header", errHeader.Error())
				return "", errFromUrl
			}
			destinationFileName = filename
		}
		if errHeader == nil {
			destinationFileName = params["filename"]
		}
	}

	if destinationFileName == "" {
		logger.Errorw("missing file name for file", "url", destinationURL, "header_value", headers.Get("Content-Disposition"), "error", utils.ErrMissingFileName.Error())
		return "", utils.ErrMissingFileName
	}

	return utils.CreateFileWithContent(body, destinationFileName, destinationPath, os.ModePerm, false, logger)
}
