package main

import (
	"fmt"
	"go_systems/src/demo2Utils"
)

func main() {
	p, err := demo2Utils.GenerateUserPassword("Sn00pyD0g!!!")
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
}
