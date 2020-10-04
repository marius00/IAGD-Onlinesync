package main

import (
	"github.com/marmyr/iagdsendmail/api/sendmail"
	"github.com/marmyr/iagdsendmail/internal/eventbus"
)

// Runs the entire application as a monolith. Useful for local testing, or deploying outside of AWS Lambda.
func main() {
	ginEngine := eventbus.MountPublicRoute(sendmail.Path, sendmail.Method, sendmail.ProcessRequest)
	ginEngine.Run()
}