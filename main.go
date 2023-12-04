package ectoinject

// import (
// 	"context"
// 	"fmt"

// 	"github.com/Gobusters/ectoinject/internal/dependency"
// )

// type TestDep interface {
// 	TestFunc(string) string
// }

// type TopStruct struct {
// 	Field1 TestDep `inject:""`
// }

// type TestStruct struct {
// 	count int
// }

// func (t *TestStruct) TestFunc(s string) string {
// 	t.count++
// 	fmt.Println("TestFunc", t.count)
// 	return s
// }

// func main() {
// 	container, err := dependency.NewDIDefaultContainer()
// 	if err != nil {
// 		panic(err)
// 	}
// 	dependency.RegisterTransient[TestDep, TestStruct](container)
// 	dependency.RegisterTransient[TopStruct, TopStruct](container)

// 	ctx := context.Background()
// 	ctx = dependency.ScopeContext(ctx)
// 	defer dependency.UnscopeContext(ctx)

// 	top1, _ := dependency.GetDependency[TopStruct](ctx)
// 	top1.Field1.TestFunc("hello")
// 	top2, _ := dependency.GetDependency[TopStruct](ctx)
// 	top2.Field1.TestFunc("hello")

// 	ctx = context.Background()
// 	ctx = dependency.ScopeContext(ctx)
// 	top3, _ := dependency.GetDependency[TopStruct](ctx)
// 	top3.Field1.TestFunc("hello")
// 	top4, _ := dependency.GetDependency[TopStruct](ctx)
// 	top4.Field1.TestFunc("hello")
// }
