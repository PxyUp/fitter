package trigger

type Trigger interface {
	Run(chan<- string)
	Stop()
}
