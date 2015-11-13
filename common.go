package main

import (
	"flag"
	"github.com/davidwalter0/k8s-bldr-api/dispatch"
)

func init() {
	flag.Parse()
	flags.ImportFlags(ExportHelper)
}

var exitOnException *bool = flag.Bool("exitOnException", false, "exit on exception")
var failureExitCode *int = flag.Int("failureExitCode", 3, "if exitOnException is enabled return to the OS this return code")
var debug *bool = flag.Bool("debug", false, "enable debug logging and other tooling")
var filename *string = flag.String("file", "unit.json", "--file json / yaml to execute tests from unit.{yaml,json}")
var verbose *bool = flag.Bool("verbose", false, "--verbose output of test progress status.")
var apiVersion *string = flag.String("apiversion", "v0.1", "configure the api version to use. only v0.1 is supported")

var flags *dispatch.ApiFlags = &dispatch.ApiFlags{}

var ExportHelper = (func() {
	flags.ExitOnException = exitOnException
	flags.FailureExitCode = failureExitCode
	flags.Debug = debug
	flags.Filename = filename
	flags.Verbose = verbose
	flags.ApiVersion = apiVersion
})
