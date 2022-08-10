package xtypes

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/simplesurance/proteus/types"
	"golang.org/x/exp/constraints"
)

// Integer is an XType for integers.
type Integer[T constraints.Integer] struct {
	DefaultValue T
	UpdateFn     func(T)
	content      struct {
		value *T
		mutex sync.Mutex
	}
}

var _ types.XType = &Integer[int]{}

// UnmarshalParam parses the input as an integer of type T.
func (d *Integer[T]) UnmarshalParam(in *string) error {
	var ptrT *T
	if in != nil {
		valT, err := parseInt[T](*in)
		if err != nil {
			return errors.New("invalid value for the numeric type")
		}

		ptrT = &valT
	}

	d.content.mutex.Lock()
	d.content.value = ptrT
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration.
func (d *Integer[T]) Value() T {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return *d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *Integer[T]) ValueValid(s string) error {
	_, err := parseInt[T](s)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *Integer[T]) GetDefaultValue() (string, error) {
	return fmt.Sprintf("%d", d.DefaultValue), nil
}

func parseInt[T constraints.Integer](v string) (ret T, _ error) {
	switch reflect.TypeOf(ret).Kind() {
	case reflect.Int:
		n, err := strconv.ParseInt(v, 10, 0)
		return T(int(n)), err
	case reflect.Int8:
		n, err := strconv.ParseInt(v, 10, 8)
		return T(int8(n)), err
	case reflect.Int16:
		n, err := strconv.ParseInt(v, 10, 16)
		return T(int16(n)), err
	case reflect.Int32:
		n, err := strconv.ParseInt(v, 10, 32)
		return T(int32(n)), err
	case reflect.Int64:
		n, err := strconv.ParseInt(v, 10, 64)
		return T(int64(n)), err

	case reflect.Uint:
		n, err := strconv.ParseUint(v, 10, 0)
		return T(uint(n)), err
	case reflect.Uint8:
		n, err := strconv.ParseUint(v, 10, 8)
		return T(uint8(n)), err
	case reflect.Uint16:
		n, err := strconv.ParseUint(v, 10, 16)
		return T(uint16(n)), err
	case reflect.Uint32:
		n, err := strconv.ParseUint(v, 10, 32)
		return T(uint32(n)), err
	case reflect.Uint64:
		n, err := strconv.ParseUint(v, 10, 64)
		return T(uint64(n)), err
	}

	panic(fmt.Errorf("unsupported type %T", ret))
}
