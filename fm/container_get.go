package fm

import (
	"fmt"
	"reflect"
)

func (c *instanceContainer) handleCandidates(instType reflect.Type, instWrappers []*instanceWrapper) (ret any, err error, handled bool) {
	if len(instWrappers) == 1 {
		if !reflect.TypeOf(instWrappers[0].inst).ConvertibleTo(instType) {
			return nil, fmt.Errorf("%w: expected %s but got %s", ErrMismatchedInstanceType, instType, reflect.TypeOf(instWrappers[0].inst)), true
		}
		return instWrappers[0].inst, nil, true
	} else if len(instWrappers) > 1 {
		return nil, ErrMultipleInstancesFound, true
	}
	return nil, nil, false
}

func (c *instanceContainer) getInstanceByParent(instType reflect.Type, name string) (any, error) {
	if c.parent != nil {
		return c.parent.getInstance(instType, name, true)
	}
	return nil, ErrInstanceNotFound
}

func (c *instanceContainer) getInstance(instType reflect.Type, name string, lockSelf bool) (any, error) {
	if lockSelf {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}

	if instType.Kind() != reflect.Interface && !(instType.Kind() == reflect.Ptr && instType.Elem().Kind() == reflect.Struct) {
		return nil, ErrMustUseInterfaceOrStructPointer
	}
	instTypeStr := instTypeString(instType)

	if name != "" {
		if ret, err, handled := c.handleCandidates(instType, c.instancesByName[name]); handled {
			return ret, err
		}
		return c.getInstanceByParent(instType, name)
	}

	if instWrappers, ok := c.instancesByType[instTypeStr]; ok && len(instWrappers) > 0 {
		if ret, err, handled := c.handleCandidates(instType, c.instancesByType[instTypeStr]); handled {
			return ret, err
		}
	}

	matchedMap := map[*instanceWrapper]int{}
	instTypeMethodNum := instType.NumMethod()
	if instTypeMethodNum != 0 {
		for i := 0; i < instTypeMethodNum; i++ {
			methodName := instType.Method(i).Name
			if instWrappers, ok := c.instancesByMethod[methodName]; ok {
				for _, inst := range instWrappers {
					matchedMap[inst]++
				}
			}
		}
		var matched []*instanceWrapper
		for instWrapper, count := range matchedMap {
			if count == instTypeMethodNum && reflect.TypeOf(instWrapper.inst).Implements(instType) {
				matched = append(matched, instWrapper)
			}
		}
		if ret, err, handled := c.handleCandidates(instType, matched); handled {
			return ret, err
		}
	}
	return c.getInstanceByParent(instType, name)
}
