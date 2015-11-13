package dispatch

import (
	"os"
	"runtime"
)

// var (
// 	info *log.Logger = logger.Info
// 	elog *log.Logger = logger.Error
// )

func trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(10, pc)
	for i := 0; i < 10; i++ {
		if pc[i] == 0 {
			break
		}
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		Info.Printf("%s:%d %s\n", file, line, f.Name())
	}
}

func RecoverWithMessage(step string, exitOnException bool, failureExitCode int) {
	if r := recover(); r != nil {
		Info.Printf("Recovered step[%s] with info %v\n", step, r)
		pc := make([]uintptr, 10) // at least 1 entry needed
		runtime.Callers(5, pc)
		f := runtime.FuncForPC(pc[1])
		file, line := f.FileLine(pc[1])
		Info.Printf("call failed at or near %s:%d %s\n", file, line, f.Name())
		if *Debug {
			trace()
		}
		if exitOnException {
			os.Exit(failureExitCode)
		}
	}
}

func check(err error, step string) {
	if err != nil {
		Elog.Println(step, err)
	}
}
