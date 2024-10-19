package fm

import "reflect"

func (c *instanceContainer) initWithDep(inst Instantiatable) error {
	deps := c.instancesPendingDep[inst]
	delete(c.instancesPendingDep, inst)
	for dep := range deps {
		err := c.initWithDep(dep)
		if err != nil {
			return err
		}
	}
	return inst.OnInstanceEvent(&InstanceEventInit{&InstanceEventCommon{instance: inst, container: c}})
}

func (c *instanceContainer) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, inst := range c.instancesPendingInit {
		instVal := reflect.ValueOf(inst)
		instTypeStr := instTypeString(instVal.Type())
		c.instancesPendingDep[inst] = map[Instantiatable]int{}
		for _, fieldIndex := range c.instanceTypeMetaMap[instTypeStr].managedFields {
			fieldVal := instVal.Elem().Field(fieldIndex)
			injectedInst, err := c.getInstance(fieldVal.Type(), "", false)
			if err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(injectedInst))
			c.instancesPendingDep[inst][injectedInst.(Instantiatable)] = 1
		}
	}

	// pre init: the fields are only injected, but the instances are not initialized
	for _, inst := range c.instancesPendingInit {
		err := inst.OnInstanceEvent(&InstanceEventPreInit{&InstanceEventCommon{instance: inst, container: c}})
		if err != nil {
			return err
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
		err := inst.OnInstanceEvent(&InstanceEventPostInit{&InstanceEventCommon{instance: inst, container: c}})
		if err != nil {
			return err
		}
	}

	c.instancesPendingInit = nil
	clear(c.instancesPendingDep)
	return nil
}
