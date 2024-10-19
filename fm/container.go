package fm

import (
	"context"
	"errors"
	"sync"
)

var ErrInstanceNotFound = errors.New("instance not found")
var ErrMultipleInstancesFound = errors.New("multiple instances found")
var ErrMismatchedInstanceType = errors.New("mismatched instance type")
var ErrMustUseInterfaceOrStructPointer = errors.New("must use interface or struct pointer get instance")

type contextKeyType struct{}

var contextKey contextKeyType

type InstanceContainer interface {
	context.Context

	Initialize() error
	NewScoped() InstanceContainer
	RegisterInstance(inst any, optionalName ...string)
}

type instanceWrapper struct {
	inst any
}

type instanceTypeMeta struct {
	managedFields map[string]int
}

type instanceContainer struct {
	context.Context

	parent *instanceContainer

	mu                  sync.RWMutex
	instanceTypeMetaMap map[string]*instanceTypeMeta
	instancesByName     map[string][]*instanceWrapper
	instancesByType     map[string][]*instanceWrapper
	instancesByMethod   map[string][]*instanceWrapper

	instancesPendingInit []Instantiatable
	instancesPendingDep  map[Instantiatable]map[Instantiatable]int
}

var _ InstanceContainer = (*instanceContainer)(nil)

func (c *instanceContainer) NewScoped() InstanceContainer {
	r := NewContainer(c.Context).(*instanceContainer)
	r.parent = c
	return r
}

func NewContainer(ctx context.Context) InstanceContainer {
	c := &instanceContainer{
		instanceTypeMetaMap: map[string]*instanceTypeMeta{},
		instancesByName:     map[string][]*instanceWrapper{},
		instancesByType:     map[string][]*instanceWrapper{},
		instancesByMethod:   map[string][]*instanceWrapper{},
		instancesPendingDep: map[Instantiatable]map[Instantiatable]int{},
	}
	c.Context = context.WithValue(ctx, contextKey, c)
	return c
}

func ContextContainer(ctx context.Context) InstanceContainer {
	c, _ := ctx.Value(contextKey).(InstanceContainer)
	return c
}
