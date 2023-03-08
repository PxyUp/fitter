package notifier

import (
	"github.com/PxyUp/fitter/pkg/parser"
)

type Notifier interface {
	Inform(result *parser.ParseResult, err error) error
}
