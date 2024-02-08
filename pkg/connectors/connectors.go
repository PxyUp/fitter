package connectors

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/builder"
)

var (
	errMaxAttempt = errors.New("reach max attempt")
	errEmpty      = errors.New("empty url")
)

type Connector interface {
	Get(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable) ([]byte, error)
}

type attemptsConnector struct {
	original Connector
	attempts uint32
}

func (r *attemptsConnector) Get(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable) ([]byte, error) {
	if r.attempts <= 0 {
		return r.original.Get(parsedValue, index, input)
	}

	for i := 0; i < int(r.attempts); i++ {
		resp, err := r.original.Get(parsedValue, index, input)
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

type nullSafe struct {
	original Connector
}

func (n *nullSafe) Get(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable) ([]byte, error) {
	resp, err := n.original.Get(parsedValue, index, input)
	if err != nil {
		return builder.NullValue.Raw(), nil
	}

	return resp, nil
}

func NullSafe(original Connector) Connector {
	return &nullSafe{
		original: original,
	}
}
