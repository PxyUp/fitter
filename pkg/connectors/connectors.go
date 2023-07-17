package connectors

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/parser/builder"
)

var (
	errMaxAttempt = errors.New("reach max attempt")
	errEmpty      = errors.New("empty url")
)

type Connector interface {
	Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error)
}

type attemptsConnector struct {
	original Connector
	attempts uint32
}

func (r *attemptsConnector) Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error) {
	if r.attempts <= 0 {
		return r.original.Get(parsedValue, index)
	}

	for i := 0; i < int(r.attempts); i++ {
		resp, err := r.original.Get(parsedValue, index)
		if err != nil || len(resp) == 0 {
			continue
		}
		return resp, nil
	}

	return nil, errMaxAttempt
}

func WithAttempts(original Connector, attempts uint32) Connector {
	return &attemptsConnector{
		original: original,
		attempts: attempts,
	}
}
