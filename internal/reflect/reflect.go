package reflect

import (
	"fmt"
	"reflect"
	"unsafe"
)

// returns the name of an interface as `modulePath.interfaceName`
func GetIntefaceName[T any]() string {
	// return the name of the interface
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	return GetReflectTypeName(interfaceType)
}

// returns the name of a reflect.Type as `modulePath.typeName`
func GetReflectTypeName(t reflect.Type) string {
	// if t is a pointer, dereference it
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	pkgPath := t.PkgPath()
	name := t.Name()

	return pkgPath + "." + name
}

// creates a new instance of a struct from a reflect.Type. t must be a struct type
func NewStructInstance(t reflect.Type) (reflect.Value, error) {
	// Ensure t is a struct type
	if t.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type '%s' is not a struct", t.Name())
	}

	// Create a new instance of the struct
	return reflect.New(t).Elem(), nil
}

// get the method of a type by name. If the type is not a pointer, get the pointer to the type and check
func GetMethodByName(t reflect.Type, name string) (reflect.Method, bool) {
	// Check for non-pointer type
	method, ok := t.MethodByName(name)
	if ok {
		return method, true
	}

	// If t is not a pointer, get the pointer to t and check
	if t.Kind() != reflect.Ptr {
		pointerType := reflect.PtrTo(t)
		method, ok = pointerType.MethodByName(name)
		return method, ok
	}

	return method, false
}

func CastType(t reflect.Type, v any) (reflect.Value, error) {
	var zeroValue reflect.Value
	vValue := reflect.ValueOf(v)
	vType := vValue.Type()

	// if v is already the correct type or implements the interface, return it
	if vType == t || (t.Kind() == reflect.Interface && vValue.Type().Implements(t)) {
		return vValue, nil
	}

	isTInterface := t.Kind() == reflect.Interface
	isTPtr := isTInterface || t.Kind() == reflect.Ptr
	isVPtr := vType.Kind() == reflect.Ptr

	// if T is a pointer or interface and v is not, create an addressable copy of v
	if isTPtr && !isVPtr {
		// Create a new value of the type of v
		newV := reflect.New(vType)

		// Set the value of v to the new value
		newV.Elem().Set(vValue)

		// Use the address of the new value
		vValue = newV
	}

	// if T is not a pointer or interface and v is, dereference v
	if !isTPtr && isVPtr {
		vValue = vValue.Elem()
	}

	// Convert v to the target type t
	if isTInterface {
		if !vValue.Type().Implements(t) {
			return zeroValue, fmt.Errorf("value of type '%s' does not implement interface '%s'", vValue.Type().String(), t.String())
		}
	} else if vValue.Type() != t {
		return zeroValue, fmt.Errorf("value of type '%s' is not of type '%s'", vValue.Type().String(), t.String())
	}

	return vValue, nil
}

func Cast[T any](v any) (T, error) {
	var zeroValue T
	tType := reflect.TypeOf((*T)(nil)).Elem()

	result, err := CastType(tType, v)
	if err != nil {
		return zeroValue, err
	}

	return result.Interface().(T), nil
}

func SetField(target reflect.Value, field reflect.StructField, value reflect.Value) error {
	// get field index
	index := field.Index

	// get the value of the field
	fieldVal := target.FieldByIndex(index)

	// can the field be set?
	canSet := fieldVal.CanSet()

	var val any // the value to set
	if value.Kind() == reflect.Ptr || !value.CanAddr() {
		val = value.Interface()
	} else {
		val = value.Addr().Interface() // prevents copying for singletons and scoped dependencies
	}

	value, err := CastType(field.Type, val)
	if err != nil {
		return err
	}

	if canSet {
		// Set the value directly if it's settable
		fieldVal.Set(value)
	} else {
		// If not settable, use unsafe to set the value
		ptr := reflect.NewAt(field.Type, unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		ptr.Set(value)
	}

	return nil
}

func GetPointerOfValue(val reflect.Value) any {
	if val.Kind() == reflect.Ptr {
		return val.Interface()
	}

	if !val.CanAddr() {
		return val.Interface()
	}

	return val.Addr().Interface()
}

func SameType[A any, B any]() bool {
	typeOfA := reflect.TypeOf((*A)(nil)).Elem()
	typeOfB := reflect.TypeOf((*B)(nil)).Elem()

	// Handle the case where A or B is an interface
	if typeOfA.Kind() == reflect.Interface {
		return typeImplementsInterface(typeOfB, typeOfA)
	}
	if typeOfB.Kind() == reflect.Interface {
		return typeImplementsInterface(typeOfA, typeOfB)
	}

	// For non-interface types, check if they are the same
	return typeOfA == typeOfB
}

// typeImplementsInterface checks if the provided type 't' implements the interface 'interfaceType'.
func typeImplementsInterface(t reflect.Type, interfaceType reflect.Type) bool {
	// Check if the type itself implements the interface
	if t.Implements(interfaceType) {
		return true
	}

	// If 't' is not a pointer type, check if a pointer to 't' implements the interface
	if t.Kind() != reflect.Ptr {
		ptrType := reflect.PtrTo(t)
		return ptrType.Implements(interfaceType)
	}

	return false
}
