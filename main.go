package main

import (
	"fmt"

	"github.com/ricjcosme/kube-monkey/config"
	"github.com/ricjcosme/kube-monkey/kubemonkey"
)

func initConfig() {
	if err := config.Init(); err != nil {
		panic(err.Error())
	}
}

func main() {
	// TODO: Set up logging

	// Initialize configs
	initConfig()

	fmt.Println("Starting kube-monkey...")

	if err := kubemonkey.Run(); err != nil {
		panic(err.Error())
	}
}
