package main

import (
	"github.com/marmyr/myservice/api/delete"
	"github.com/marmyr/myservice/api/download"
	"github.com/marmyr/myservice/api/migrate"
	"github.com/marmyr/myservice/api/remove"
	"github.com/marmyr/myservice/api/session/auth"
	"github.com/marmyr/myservice/api/session/login"
	"github.com/marmyr/myservice/api/session/logincheck"
	"github.com/marmyr/myservice/api/session/logout"
	"github.com/marmyr/myservice/api/upload"
	"github.com/marmyr/myservice/internal/routing"
	"log"
)

// Runs the entire application as a single application. Useful for local testing, or deploying outside of AWS Lambda.
func main() {
	ginEngine := routing.Build()
	routing.AddProtectedRoute(ginEngine, delete.Path, delete.Method, delete.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, download.Path, download.Method, download.ProcessRequest)
	routing.AddPublicRoute(ginEngine, migrate.Path, migrate.Method, migrate.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, remove.Path, remove.Method, remove.ProcessRequest)
	routing.AddPublicRoute(ginEngine, auth.Path, auth.Method, auth.ProcessRequest)
	routing.AddPublicRoute(ginEngine, login.Path, login.Method, login.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logincheck.Path, logincheck.Method, logincheck.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logout.Path, logout.Method, logout.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, upload.Path, upload.Method, upload.ProcessRequest)

	if err := ginEngine.Run(); err != nil {
		log.Printf("Error starting gin %v", err)
	}
	log.Printf("I guess that was that.")
}