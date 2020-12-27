package main

import (
	"fmt"
	"github.com/marmyr/myservice/internal/storage"
)

// Runs maintenance work such as deleting failed authentication attempts and throttle entries
func main() {
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