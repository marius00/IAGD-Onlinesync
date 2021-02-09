package main

import (
	"github.com/marmyr/iagdbackup/api/buddyitems"
	"github.com/marmyr/iagdbackup/api/character"
	"github.com/marmyr/iagdbackup/api/delete"
	"github.com/marmyr/iagdbackup/api/download"
	"github.com/marmyr/iagdbackup/api/getbuddyid"
	"github.com/marmyr/iagdbackup/api/migrate"
	"github.com/marmyr/iagdbackup/api/remove"
	"github.com/marmyr/iagdbackup/api/search"
	"github.com/marmyr/iagdbackup/api/session/auth"
	"github.com/marmyr/iagdbackup/api/session/login"
	"github.com/marmyr/iagdbackup/api/session/logincheck"
	"github.com/marmyr/iagdbackup/api/session/logout"
	"github.com/marmyr/iagdbackup/api/upload"
	"github.com/marmyr/iagdbackup/internal/routing"
	"log"
)

// Runs the entire application as a single application. Useful for local testing, or deploying outside of AWS Lambda.
func main() {
	ginEngine := routing.Build()
	routing.AddPublicRoute(ginEngine, search.Path, search.Method, search.ProcessRequest)
	routing.AddPublicRoute(ginEngine, buddyitems.Path, buddyitems.Method, buddyitems.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, delete.Path, delete.Method, delete.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, download.Path, download.Method, download.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, getbuddyid.Path, getbuddyid.Method, getbuddyid.ProcessRequest)
	routing.AddPublicRoute(ginEngine, migrate.Path, migrate.Method, migrate.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, remove.Path, remove.Method, remove.ProcessRequest)
	routing.AddPublicRoute(ginEngine, auth.Path, auth.Method, auth.ProcessRequest)
	routing.AddPublicRoute(ginEngine, login.Path, login.Method, login.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logincheck.Path, logincheck.Method, logincheck.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logout.Path, logout.Method, logout.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, upload.Path, upload.Method, upload.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.UploadPath, character.UploadMethod, character.UploadProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.DownloadPath, character.DownloadMethod, character.DownloadProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.ListPath, character.ListMethod, character.ListProcessRequest)

	if err := ginEngine.Run(); err != nil {
		log.Printf("Error starting gin %v", err)
	}
	log.Printf("I guess that was that.")
}