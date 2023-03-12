package log

import (
	"fmt"
	"github.com/isxuelinme/poe_unoffical_api/internal/godump"
	"os"
	"runtime/debug"
	"strings"
	"time"
	//"xuelin.me/go/lib/godump"
)

type LogSaveFunc func(log string)

var logSaveFunc LogSaveFunc

type LogShowAllFunc func(date string)

var logShowAllFunc LogShowAllFunc

func INIT(save LogSaveFunc, show LogShowAllFunc) {
	logSaveFunc = save
	logShowAllFunc = show
}
func LoggerOut(title string, out []interface{}, outputStack bool) {
	stackStr := ""
	outStack := ""
	if outputStack {
		stack := debug.Stack()
		stackSlice := strings.Split(string(stack), "\n")
		stackSlice = stackSlice[7 : len(stackSlice)-1]
		for _, stack := range stackSlice {
			stackStr += stack + "\n"
		}
		outStr := ""
		if len(out) > 1 {
			for i, val := range out {
				if i == 0 {
					outStr += fmt.Sprintf("%s", val) + "\n"
				} else {
					outStr += godump.StringsDump(val)
				}
			}
		} else {
			outStr += fmt.Sprintf("%s", out)
		}
		outStack = fmt.Sprintf("[%s] %s %s\n---stacks---\n%s", title, time.Now().Format("2006/01/02 15:04:05"), outStr, stackStr)
	} else {
		outStr := ""
		if len(out) > 1 {
			for i, val := range out {
				if i == 0 {
					outStr += fmt.Sprintf("%s", val) + "\n"
				} else {
					outStr += godump.StringsDump(val)
				}
			}
		} else {
			outStr += fmt.Sprintf("%s", out)
		}
		stack := debug.Stack()
		stackSlice := strings.Split(string(stack), "\n")
		stackSlice = stackSlice[7:9]
		stackStr := "\n"
		for _, stack := range stackSlice {
			stackStr += stack + "\n"
		}
		outStack = fmt.Sprintf("[%s] %s %s %s", title, time.Now().Format("2006/01/02 15:04:05"), outStr, stackStr)
	}

	fmt.Println(outStack)
	if logShowAllFunc == nil {
		return
	}
	logSaveFunc(outStack)
}

func Error(out ...interface{}) {
	LoggerOut("Error", out, true)
}
func Info(out ...interface{}) {
	LoggerOut("Info", out, false)
}
func Panic(out ...interface{}) {
	LoggerOut("Panic", out, true)
}

func PanicAndStopWorld(out ...interface{}) {
	LoggerOut("Panic", out, true)
	os.Exit(0)
}

func ShowLog(date string) {
	if logShowAllFunc == nil {
		return
	}
	logShowAllFunc(date)
}
