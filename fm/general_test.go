package fm_test

import (
	"context"
	"github.com/go-farmyard/farmyard/fm"
	"github.com/stretchr/testify/assert"
	"testing"
)

type intf1 interface {
	Intf1Func() int
}

type impl1 struct {
	val int

	fm.AutoInitialize

	Impl2 intf2
}

func (i *impl1) Intf1Func() int {
	return i.val * 10
}

type intf2 interface {
	Intf2Func() int
}

type impl2 struct {
	val int

	fm.AutoInitialize

	Impl1 intf1
}

func (i *impl2) Intf2Func() int {
	return i.val * 10
}

func TestGeneral(t *testing.T) {
	c := fm.NewContainer(context.Background())

	c.RegisterInstance(&impl1{val: 1})
	c.RegisterInstance(impl2{val: 2}, "name2") // struct will be converted to pointer internally

	assert.NoError(t, c.Initialize())

	assert.Equal(t, 1, fm.RequirePointer[impl1](c).val)
	assert.Equal(t, 10, fm.RequireInterface[intf1](c).Intf1Func())

	assert.Equal(t, 2, fm.RequirePointer[impl2](c).val)
	assert.Equal(t, 2, fm.RequirePointer[impl2](c, "name2").val)

	assert.Equal(t, 20, fm.RequirePointer[impl1](c).Impl2.Intf2Func())
	assert.Equal(t, 10, fm.RequirePointer[impl2](c).Impl1.Intf1Func())

	// the correct way is to use "GetInstance[*impl1]" or RequirePointer, because all instances are stored as pointers
	_, err := fm.GetInstance[impl1](c)
	assert.ErrorIs(t, err, fm.ErrMustUseInterfaceOrStructPointer)
}
