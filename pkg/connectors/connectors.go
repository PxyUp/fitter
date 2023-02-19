package connectors

import "errors"

var (
	errMaxAttempt = errors.New("reach max attempt")
)

type Connector interface {
	Get() ([]byte, error)
}

type attemptsConnector struct {
	original Connector
	attempts uint32
}

func (r *attemptsConnector) Get() ([]byte, error) {
	if r.attempts <= 0 {
		return r.original.Get()
	}

	for i := 0; i < int(r.attempts); i++ {
		resp, err := r.original.Get()
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
