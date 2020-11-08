package main

import (
	"github.com/marmyr/myservice/api/save"
	"github.com/marmyr/myservice/internal/eventbus"
	"log"
)

// Runs the entire application as a monolith. Useful for local testing, or deploying outside of AWS Lambda.
func main() {
	ginEngine := eventbus.MountPublicRoute(save.Path, save.Method, save.ProcessRequest)
	// eventbus.AddPublicRoute(ginEngine, save.Path, save.Method, save.ProcessRequest)
	ginEngine.Run()
	log.Printf("I guess that was that.")
}