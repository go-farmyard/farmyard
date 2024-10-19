package fm

import (
	"reflect"
)

type AutoInitialize struct{}

type InstancePreInitializer interface {
	OnInstancePreInit() error
}

type InstanceInitializer interface {
	OnInstanceInit() error
}

type InstancePostInitializer interface {
	OnInstancePostInit() error
}

func isInstanceInitializable(inst any) bool {
	_, ok1 := inst.(InstancePreInitializer)
	_, ok2 := inst.(InstanceInitializer)
	_, ok3 := inst.(InstancePostInitializer)
	return ok1 || ok2 || ok3
}

func (c *instanceContainer) initWithDep(inst any) error {
	deps := c.instancesPendingDep[inst]
	delete(c.instancesPendingDep, inst)
	for dep := range deps {
		err := c.initWithDep(dep)
		if err != nil {
			return err
		}
	}
	if init, ok := inst.(InstanceInitializer); ok {
		return init.OnInstanceInit()
	}
	return nil
}

func (c *instanceContainer) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, inst := range c.instancesPendingInit {
		instVal := reflect.ValueOf(inst)
		instTypeStr := instTypeString(instVal.Type())
		c.instancesPendingDep[inst] = map[any]int{}
		for _, fieldIndex := range c.instanceTypeMetaMap[instTypeStr].managedFields {
			fieldVal := instVal.Elem().Field(fieldIndex)
			injectedInst, err := c.getInstance(fieldVal.Type(), "", false)
			if err != nil {
				return err
			}

			injectedInstValue := reflect.ValueOf(injectedInst)
			fieldVal.Set(injectedInstValue)
			injectedInstTypeStr := instTypeString(injectedInstValue.Type())
			if injectedInstTypeMeta, ok := c.instanceTypeMetaMap[injectedInstTypeStr]; ok && injectedInstTypeMeta.isManaged {
				c.instancesPendingDep[inst][injectedInst] = 1
			}
		}
	}

	// pre init: the fields are only injected, but the instances are not initialized
	for _, inst := range c.instancesPendingInit {
		if inst, ok := inst.(InstancePreInitializer); ok {
			err := inst.OnInstancePreInit()
			if err != nil {
				return err
			}
		}
	}

	// init: the instances are being initialized by the dependency order
	for len(c.instancesPendingDep) != 0 {
		for inst := range c.instancesPendingDep {
			err := c.initWithDep(inst)
			if err != nil {
				return err
			}
			break
		}
	}

	// post init: all the instances have been initialized
	for _, inst := range c.instancesPendingInit {
		if inst, ok := inst.(InstancePostInitializer); ok {
			err := inst.OnInstancePostInit()
			if err != nil {
				return err
			}
		}
	}

	c.instancesPendingInit = nil
	clear(c.instancesPendingDep)
	return nil
}
