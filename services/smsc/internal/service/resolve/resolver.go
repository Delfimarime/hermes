package resolve

type ValueResolver interface {
	Get(name string) (string, error)
}
