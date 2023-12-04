package reflect

import (
	"fmt"
	"reflect"
)

// returns the name of an interface as `modulePath.interfaceName`
func GetIntefaceName[T any]() string {
	// return the name of the interface
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	return GetReflectTypeName(interfaceType)
}

// returns the name of a reflect.Type as `modulePath.typeName`
func GetReflectTypeName(t reflect.Type) string {
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
