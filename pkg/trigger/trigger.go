package trigger

import "github.com/PxyUp/fitter/pkg/builder"

type Message struct {
	Name  string
	Value builder.Interfacable
}

type Trigger interface {
	Run(chan<- *Message)
	Stop()
}
