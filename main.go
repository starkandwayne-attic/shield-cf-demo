package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jhunt/vcaptive"
)

type Data struct {
	System       string `json:"system"`
	Summary      string `json:"summary"`
	Verification string `json:"verification"`
	OK bool `json:"ok"`
}

type System interface {
	Configure(vcaptive.Services) (bool, error)
	Setup() error
	Teardown() error
	Summarize() Data
}

var Systems map[string]System

func main() {
	if os.Getenv("VCAP_SERVICES") == "" {
		fmt.Fprintf(os.Stderr, "No $VCAP_SERVICES environment variable found; are you running me under Cloud Foundry?\n")
		os.Exit(1)
	}
	if os.Getenv("PORT") == "" {
		fmt.Fprintf(os.Stderr, "No $PORT environment variable found; are you running me under Cloud Foundry?\n")
		os.Exit(1)
	}

	services, err := vcaptive.ParseServices(os.Getenv("VCAP_SERVICES"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: %s\n", err)
		os.Exit(1)
	}

	var (
		name string
		system System
		found bool
	)
	/* loop over our known Systems and find the first one that matches */
	found = false
	for name, system = range Systems {
		fmt.Fprintf(os.Stdout, "trying to configure the '%s' system...\n", name)
		ok, err := system.Configure(services)
		if !ok {
			continue
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to configure %s system: %s\n", name, err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "configured the '%s' system successfully, resuming normal startup...\n", name)
		found = true
		break
	}
	if !found {
		fmt.Fprintf(os.Stderr, "No suitable services are bound to this instance of the SHIELD demo application; please try binding a marketplace service and restarting this appplication.\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stdout, "setting up data for the %s system...\n", name)
	if err := system.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set up the %s system: %s\n", name, err)
		os.Exit(3)
	}

	fmt.Fprintf(os.Stdout, "setting up the /data API handler...\n")
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			summary := system.Summarize()
			b, err := json.Marshal(summary)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "an internal error has occurred")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			fmt.Fprintf(w, "%s\n", string(b))

		case "DELETE":
			if err := system.Teardown(); err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "%s\n", err)
				return
			}

			w.WriteHeader(200)
			fmt.Fprintf(w, "deleted\n")

		default:
			w.WriteHeader(405)
		}
	})

	fmt.Fprintf(os.Stdout, "setting up the / asset handler...\n")
	http.Handle("/", http.FileServer(http.Dir("htdocs")))

	fmt.Fprintf(os.Stdout, "listening on :%s\n", os.Getenv("PORT"))
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
}
