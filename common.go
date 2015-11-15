package main

import (
	"flag"
	"fmt"
	"github.com/davidwalter0/k8s-bldr-api/dispatch"
	"os"
	"strings"
)

func init() {
	flag.Parse()
	flags.ImportFlags(ExportHelper)
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	if *version {
		fmt.Println(me, "version built as:", Build, "commit:", Commit)
		os.Exit(0)
	}
	fmt.Println(me, "version built as:", Build, "commit:", Commit)
}

var exitOnException *bool = flag.Bool("exitOnException", false, "exit on exception")
var failureExitCode *int = flag.Int("failureExitCode", 3, "if exitOnException is enabled return to the OS this return code")
var debug *bool = flag.Bool("debug", false, "enable debug logging and other tooling")
var filename *string = flag.String("file", "unit.json", "--file json / yaml to execute tests from unit.{yaml,json}")
var verbose *bool = flag.Bool("verbose", false, "--verbose output of test progress status.")
var apiVersion *string = flag.String("apiversion", "v0.1", "configure the api version to use. only v0.1 is supported")
var version *bool = flag.Bool("version", false, "print the build text and hash of commit then exit")
var port *uint = flag.Uint("port", 9999, "configure the port on which to listen")
var Build string
var Commit string

var flags *dispatch.ApiFlags = &dispatch.ApiFlags{}

var ExportHelper = (func() {
	flags.ExitOnException = exitOnException
	flags.FailureExitCode = failureExitCode
	flags.Debug = debug
	flags.Filename = filename
	flags.Verbose = verbose
	flags.ApiVersion = apiVersion
	flags.Port = port
})
