package main

import (
	"dispatcher/registrant"
	"dispatcher/server"
	"dispatcher/src/department"
	"fmt"
)

func main() {

	fmt.Println("Hello, this is official transaction framework")
	dispatcher := registrant.NewRegisterDispatch()
	department.NewProductDepartment()
	server.InitServer(dispatcher)
}
