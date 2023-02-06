package connectors

type Connector interface {
	Get() ([]byte, error)
}
