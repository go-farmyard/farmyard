package fm

type AutoInstantiate struct{}

var _ Instantiatable = (*AutoInstantiate)(nil)

func (d *AutoInstantiate) OnInstanceEvent(e InstanceEvent) error {
	return nil
}

type CustomInstantiate struct{}
