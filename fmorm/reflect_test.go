package fmorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTableName(t *testing.T) {
	var rows []*testModel
	name, _ := getTableFromSlice(&rows)
	assert.Equal(t, "test_table", name)
}

func TestResetToZeroValue(t *testing.T) {
	v := &testModel{
		ID: 1,
		K1: 2,
	}
	resetToZeroValue(v)
	assert.EqualValues(t, 0, v.ID)
	assert.EqualValues(t, 0, v.K1)
}
