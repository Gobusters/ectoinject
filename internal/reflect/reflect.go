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
	var zeroValue reflect.Value // Zero value of T

	vType := reflect.TypeOf(v)

	// if v is already the correct type, return it
	if vType == t {
		return reflect.ValueOf(v), nil
	}

	isTPtr := t.Kind() == reflect.Ptr
	isVPtr := vType.Kind() == reflect.Ptr

	// if T is a pointer and v is not, get the address of v
	if isTPtr && !isVPtr {
		rv := reflect.ValueOf(v)
		if rv.CanAddr() {
			v = rv.Addr().Interface()
		} else {
			// Cannot take the address of v
			return zeroValue, fmt.Errorf("cannot take address of the provided value '%s' to convert to %s", vType.Name(), t.Name())
		}
	}

	// if T is not a pointer and v is, dereference v
	if !isTPtr && isVPtr {
		v = reflect.ValueOf(v).Elem().Interface()
	}

	result := reflect.ValueOf(v)
	if result.Type() != t {
		// refetch the type
		vType = reflect.TypeOf(v)
		return zeroValue, fmt.Errorf("value of type '%s' is not of type '%s'", vType.Name(), t.Name())
	}

	return result, nil
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

func CastValue[T any](v reflect.Value) (T, error) {
	var zeroValue T

	// if v is already the correct type, return it
	result, ok := v.Interface().(T)
	if ok {
		return result, nil
	}

	// if v is not the correct type, cast it
	tType := reflect.TypeOf((*T)(nil)).Elem()
	castV, err := CastType(tType, v.Interface())
	if err != nil {
		return zeroValue, err
	}

	return castV.Interface().(T), nil
}

func SetField(target reflect.Value, field reflect.StructField, value reflect.Value) error {
	// get field index
	index := field.Index

	// get the value of the field
	fieldVal := target.FieldByIndex(index)

	// is the field an interface?
	isFieldInterface := field.Type.Kind() == reflect.Interface

	// is the field a pointer?
	isFieldPtrOrInterface := isFieldInterface || field.Type.Kind() == reflect.Ptr

	// is the value a pointer?
	isValuePtr := value.Kind() == reflect.Ptr

	// can the field be set?
	canSet := fieldVal.CanSet()

	// if the field is a pointer and the value is not, get the address of the value
	if isFieldPtrOrInterface && !isValuePtr {
		rv := reflect.ValueOf(value)
		if !rv.CanAddr() {
			return fmt.Errorf("cannot take address of the provided value '%s' to set to field '%s'", value.Type().Name(), field.Name)
		}

		value = rv.Addr()
	}

	// if the field is not a pointer and the value is, dereference the value
	if !isFieldPtrOrInterface && isValuePtr {
		value = value.Elem()
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
