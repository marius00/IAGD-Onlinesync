package main

import (
	"github.com/marmyr/myservice/api/delete"
	"github.com/marmyr/myservice/api/download"
	"github.com/marmyr/myservice/api/migrate"
	"github.com/marmyr/myservice/api/remove"
	"github.com/marmyr/myservice/api/session/auth"
	"github.com/marmyr/myservice/api/session/login"
	"github.com/marmyr/myservice/api/session/logincheck"
	"github.com/marmyr/myservice/api/upload"
	"github.com/marmyr/myservice/internal/eventbus"
	"log"
)

// Runs the entire application as a single application. Useful for local testing, or deploying outside of AWS Lambda.
func main() {
	ginEngine := eventbus.Build()
	eventbus.AddProtectedRoute(ginEngine, delete.Path, delete.Method, delete.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, download.Path, download.Method, download.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, migrate.Path, migrate.Method, migrate.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, remove.Path, remove.Method, remove.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, auth.Path, auth.Method, auth.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, login.Path, login.Method, login.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, logincheck.Path, logincheck.Method, logincheck.ProcessRequest)
	eventbus.AddProtectedRoute(ginEngine, upload.Path, upload.Method, upload.ProcessRequest)

	if err := ginEngine.Run(); err != nil {
		log.Printf("Error starting gin %v", err)
	}
	log.Printf("I guess that was that.")
}