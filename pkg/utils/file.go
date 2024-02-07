package utils

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/logger"
	"os"
	"path"
)

var (
	ErrMissingFileName = errors.New("missing file name")
)

func CreateFileWithContent(content []byte, destinationFileName string, destinationPath string, mode os.FileMode, append bool, logger logger.Logger) (string, error) {
	if destinationFileName == "" {
		return "", ErrMissingFileName
	}

	localDest := path.Join(destinationPath, destinationFileName)
	logger.Debugw("storing file", "path", localDest)
	if _, errDir := os.Stat(destinationPath); os.IsNotExist(errDir) {
		errCreationOfDir := os.MkdirAll(destinationPath, os.ModePerm)
		if errCreationOfDir != nil {
			logger.Errorw("unable to create directory", "path", destinationPath, "error", errCreationOfDir.Error())
			return "", errCreationOfDir
		}
	}

	if append {
		file, errFile := os.OpenFile(localDest, os.O_APPEND|os.O_WRONLY|os.O_CREATE, mode)
		if errFile != nil {
			logger.Errorw("unable to open local file", "dest", localDest, "error", errFile.Error())
			return "", errFile
		}

		defer func() {
			_ = file.Close()
		}()

		_, errWrite := file.Write(content)
		if errWrite != nil {
			logger.Errorw("unable to append content to local file", "dest", localDest, "error", errWrite.Error())
			return "", errWrite
		}

		return localDest, nil
	}

	err := os.WriteFile(localDest, content, mode)
	if err != nil {
		logger.Errorw("unable to write content to local file", "dest", localDest, "error", err.Error())
		return "", err
	}

	return localDest, nil
}
