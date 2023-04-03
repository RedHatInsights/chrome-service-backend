package main

import (
	"fmt"
	"path/filepath"

	"github.com/gookit/validate"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func handleValidationError(res *validate.Validation, moduleId string) {

	fmt.Println("Validation faile for module: ", moduleId)
	fmt.Println(res.Errors)               // all error messages
	fmt.Println(res.Errors.Field("Name")) // returns error messages of the field

	panic(fmt.Sprintf("Validation faile for module: %s", moduleId))
}

func main() {
	cwd, err := filepath.Abs(".")
	handleErr(err)
	fmt.Println("Validating static schemas")
	fmt.Println("Validating module definitions")

	err = validateModules(cwd)
	handleErr(err)
}
