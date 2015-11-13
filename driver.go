package main

import (
	"github.com/davidwalter0/k8s-bldr-api/dispatch"
	"os"
)

func main() {
	exit, _, response := dispatch.Dispatch(*dispatch.Filename, *dispatch.Debug, *dispatch.Verbose)
	if *dispatch.Debug {
		dispatch.Info.Println(response)
	}
	os.Exit(exit)
}
