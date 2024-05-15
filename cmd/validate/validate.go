package main

import (
	"fmt"
	"path/filepath"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	cwd, err := filepath.Abs(".")
	handleErr(err)
	fmt.Println("Validating chrome schemas")
	fmt.Println("Validating module definitions")
	validateModules(cwd)
	fmt.Println("Validating navigation definitions")
	validateNavigation(cwd)
	fmt.Println("Validating section definitions")
	validateServices(cwd)
	fmt.Println("Validating widget definitions")
	validateDashboardDefaults(cwd)
}
