package reflect

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStructInstance(t *testing.T) {
	type testStruct struct {
	}

	// create a new instance of testStruct
	instance, err := NewStructInstance(reflect.TypeOf(testStruct{}))
	assert.NoError(t, err)

	// check if the instance is of type testStruct
	if _, ok := instance.Interface().(testStruct); !ok {
		t.Errorf("instance is not of type testStruct")
	}
}

type TestDep interface {
	GetString() string
	GetNumber() int
	IncrementCount() int
}

type TestDep1 struct {
	Text  string
	Num   int
	Count int
}

func (t *TestDep1) Constructor() *TestDep1 {
	t.Text = "test"
	t.Num = 0
	t.Count = 0
	return t
}

func (t *TestDep1) GetString() string {
	return t.Text
}

func (t *TestDep1) GetNumber() int {
	return t.Num
}

func (t *TestDep1) IncrementCount() int {
	t.Count++
	return t.Count
}

func TestCastType(t *testing.T) {
	// cast the instance to a testStruct
	tType := reflect.TypeOf((*TestDep)(nil)).Elem()
	castInstance, err := CastType(tType, TestDep1{Text: "test", Num: 0, Count: 0})
	assert.NoError(t, err)

	// check if the instance is of type testStruct
	if _, ok := castInstance.Interface().(TestDep); !ok {
		t.Errorf("instance is not of type testStruct")
	}

	assert.Equal(t, "test", castInstance.Interface().(TestDep).GetString())
}
