package main

import (
	"fmt"

	"github.com/ricjcosme/kube-monkey/config"
	"github.com/ricjcosme/kube-monkey/kubemonkey"
	"github.com/ricjcosme/kube-monkey/healthandmetrics"
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

	// Initialize health checker and metrics
	fmt.Println("Starting health checker and metrics...")
	go healthandmetrics.Run()

	fmt.Println("Starting kube-monkey...")

	if err := kubemonkey.Run(); err != nil {
		panic(err.Error())
	}

}
