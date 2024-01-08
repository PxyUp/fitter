package parser

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/PxyUp/fitter/pkg/utils"
	"mime"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

var (
	errMissingFileName = errors.New("missing file name")
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

func ProcessFileField(parsedValue builder.Jsonable, index *uint32, field *config.FileFieldConfig, logger logger.Logger) (string, error) {
	destinationFileName := utils.Format(field.FileName, parsedValue, index)
	destinationPath := utils.Format(field.Path, parsedValue, index)
	destinationURL := utils.Format(field.Url, parsedValue, index)

	connector := connectors.NewAPI(destinationURL, field.Config, http_client.Client).WithLogger(logger.With("connector", "file"))

	headers, body, err := connector.GetWithHeaders(parsedValue, index)
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
		logger.Errorw("missing file name for file", "url", destinationURL, "error", "header_value", headers.Get("Content-Disposition"), errMissingFileName.Error())
		return "", errMissingFileName
	}

	localDest := path.Join(destinationPath, destinationFileName)
	logger.Debugw("storing file", "path", localDest)
	if _, errDir := os.Stat(destinationPath); os.IsNotExist(errDir) {
		errCreationOfDir := os.Mkdir(destinationPath, os.ModePerm)
		if errCreationOfDir != nil {
			logger.Errorw("unable to create directory", "path", destinationPath, "error", errCreationOfDir.Error())
			return "", errCreationOfDir
		}
	}

	err = os.WriteFile(localDest, body, os.ModePerm)
	if err != nil {
		logger.Errorw("unable to write content to local file", "dest", localDest, "error", err.Error())
		return "", err
	}

	return localDest, nil
}
