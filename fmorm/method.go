package fmorm

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"reflect"
)

func OneWhere[T TableNameProvider](d *Orm, clause any, args ...any) (ret T) {
	typ := reflect.TypeOf(ret)
	model := reflect.New(typ.Elem()).Interface().(T)
	found, err := d.Where(clause, args...).SelectOne(model)
	fmutil.MustNoError(err, "sql select failed")
	if !found {
		return
	}
	return model
}

func OneByModel[T TableNameProvider](d *Orm, model T) (_ T) {
	typ := reflect.TypeOf(model)
	res := reflect.New(typ.Elem()).Interface().(T)
	found, err := d.Where(CondModel(model)).SelectOne(res)
	fmutil.MustNoError(err, "sql select failed")
	if !found {
		return
	}
	return res
}

func One[T any](cb *CmdBuilder) (ret T) {
	var found bool
	var err error
	v := reflect.TypeOf(ret)
	if v.Kind() == reflect.Ptr {
		typ := reflect.TypeOf(ret)
		ret = reflect.New(typ.Elem()).Interface().(T)
		found, err = cb.SelectOne(ret)
	} else {
		found, err = cb.SelectOne(&ret)
	}
	fmutil.MustNoError(err, "sql select failed")
	if !found {
		return
	}
	return ret
}

func All[T any](cb *CmdBuilder) (ret []*T) {
	err := cb.SelectAll(&ret)
	fmutil.MustNoError(err, "sql select failed")
	return ret
}

func DeleteByModel(d *Orm, model TableNameProvider) int64 {
	return DeleteWhere(d, model, CondModel(model))
}

func DeleteWhere(d *Orm, model TableNameProvider, clause any, args ...any) int64 {
	res, err := d.Table(model).Where(clause, args...).Delete()
	fmutil.MustNoError(err, "sql delete failed")
	n, err := res.RowsAffected()
	fmutil.MustNoError(err, "sql delete failed")
	return n
}

func UpdateModelByModel(d *Orm, model TableNameProvider, where TableNameProvider) int64 {
	return UpdateModelWhere(d, model, CondModel(where))
}

func UpdateModelWhere(d *Orm, model TableNameProvider, clause any, args ...any) int64 {
	res, err := d.Where(clause, args...).UpdateModel(model)
	fmutil.MustNoError(err, "sql update failed")
	n, err := res.RowsAffected()
	fmutil.MustNoError(err, "sql update failed")
	return n
}

func InsertModelGetId(d *Orm, model TableNameProvider) int64 {
	res, err := d.Cmd().InsertModel(model)
	fmutil.MustNoError(err, "sql insert failed")
	n, err := res.LastInsertId()
	fmutil.MustNoError(err, "sql insert failed")
	return n
}
