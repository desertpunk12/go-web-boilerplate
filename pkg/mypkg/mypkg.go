package mypkg

import "fmt"

func Test() {
	testVar, err := MyTestFunc()
	if err != nil {
		panic(err)
	}
	fmt.Printf("my package test! : %s", testVar)
}

func MyTestFunc() (string, error) {
	return "passed!", nil
}
