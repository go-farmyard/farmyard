package fm

type InstanceEvent interface {
	GetInstance() Instantiatable
	GetContainer() InstanceContainer
}

type InstanceEventCommon struct {
	instance  Instantiatable
	container InstanceContainer
}

func (i *InstanceEventCommon) GetInstance() Instantiatable {
	return i.instance
}

func (i *InstanceEventCommon) GetContainer() InstanceContainer {
	return i.container
}

type InstanceEventPreInit struct {
	*InstanceEventCommon
}

type InstanceEventInit struct {
	*InstanceEventCommon
}

type InstanceEventPostInit struct {
	*InstanceEventCommon
}

type InstanceEventAllReady struct {
	*InstanceEventCommon
}

type Instantiatable interface {
	OnInstanceEvent(e InstanceEvent) error
}
