package trigger

import "github.com/PxyUp/fitter/pkg/builder"

type Message struct {
	Name  string
	Value builder.Jsonable
}

type Trigger interface {
	Run(chan<- Message)
	Stop()
}
