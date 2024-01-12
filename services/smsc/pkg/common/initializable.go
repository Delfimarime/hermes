package common

type PostConfigurable interface {
	AfterPropertiesSet() error
}
