package main

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/marmyr/iagdbackup/api"
	"github.com/marmyr/iagdbackup/api/buddyitems"
	"github.com/marmyr/iagdbackup/api/character"
	"github.com/marmyr/iagdbackup/api/delete"
	"github.com/marmyr/iagdbackup/api/download"
	"github.com/marmyr/iagdbackup/api/getbuddyid"
	"github.com/marmyr/iagdbackup/api/migrate"
	"github.com/marmyr/iagdbackup/api/remove"
	"github.com/marmyr/iagdbackup/api/session/auth"
	"github.com/marmyr/iagdbackup/api/session/authstatus"
	"github.com/marmyr/iagdbackup/api/session/login"
	"github.com/marmyr/iagdbackup/api/session/logincheck"
	"github.com/marmyr/iagdbackup/api/session/logout"
	"github.com/marmyr/iagdbackup/api/upload"
	"github.com/marmyr/iagdbackup/api/ws"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/wshub"
	"log"
	"time"
)

// Runs the entire application as a single process. This is the sole entrypoint,
// deployed as a Docker image on Coolify with /storage mounted as a persistent volume.
func main() {
	ginEngine := routing.Build()
	routing.AddPublicRoute(ginEngine, buddyitems.Path, buddyitems.Method, buddyitems.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, delete.Path, delete.Method, delete.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, download.Path, download.Method, download.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, getbuddyid.Path, getbuddyid.Method, getbuddyid.ProcessRequest)
	routing.AddPublicRoute(ginEngine, migrate.Path, migrate.Method, migrate.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, remove.Path, remove.Method, remove.ProcessRequest)
	routing.AddPublicRoute(ginEngine, auth.Path, auth.Method, auth.ProcessRequest)
	routing.AddPublicRoute(ginEngine, login.Path, login.Method, login.ProcessRequest)
	routing.AddPublicRoute(ginEngine, authstatus.Path, authstatus.Method, authstatus.ProcessRequest)
	routing.AddPublicRoute(ginEngine, healthcheck.Path, healthcheck.Method, healthcheck.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logincheck.Path, logincheck.Method, logincheck.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, logout.Path, logout.Method, logout.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, upload.Path, upload.Method, upload.ProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.UploadPath, character.UploadMethod, character.UploadProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.DownloadPath, character.DownloadMethod, character.DownloadProcessRequest)
	routing.AddProtectedRoute(ginEngine, character.ListPath, character.ListMethod, character.ListProcessRequest)

	// Live sync: relays item additions/deletions between a user's machines.
	hub := wshub.New()
	routing.AddProtectedRoute(ginEngine, ws.Path, ws.Method, ws.ProcessRequest(hub))

	// Seed core.db (user directory + records) from the read-only MySQL source.
	// No-op once MySQL is decommissioned.
	if err := storage.BootstrapFromMySQL(); err != nil {
		log.Fatalf("Error bootstrapping core.db from MySQL, %v", err)
	}

	if err := storage.Preload(); err != nil {
		log.Fatalf("Error preloading record cache, %v", err)
	}

	if err := storage.PreloadMigrationState(); err != nil {
		log.Fatalf("Error preloading migration state, %v", err)
	}

	s := gocron.NewScheduler(time.UTC)
	s.Every(12).Hours().Do(maintenance)
	s.StartAsync()

	// Continuously and slowly drain users off MySQL so the legacy database can be
	// decommissioned without waiting for every user to log in. Throttled to spare
	// the host; runs for the lifetime of the process.
	go storage.RunBackgroundDrain(2*time.Second, 1*time.Hour)

	if err := ginEngine.Run(); err != nil {
		log.Printf("Error starting gin %v", err)
	}

	log.Printf("I guess that was that.")
}

// Runs maintenance work such as deleting failed authentication attempts and throttle entries
func maintenance() {
	fmt.Println("Starting maintenance")

	itemDb := storage.ItemDb{}
	if err := itemDb.Maintenance(); err != nil {
		fmt.Printf("Error performing maintenance on item db, %v", err)
	}

	throttleDb := storage.ThrottleDb{}
	if err := throttleDb.Maintenance(); err != nil {
		fmt.Printf("Error performing maintenance on throttle db, %v", err)
	}

	authDb := storage.AuthDb{}
	if err := authDb.Maintenance(); err != nil {
		fmt.Printf("Error performing maintenance on auth db, %v", err)
	}

	fmt.Println("Maintenance finished")
}
