package recover

import (
	"github.com/isxuelinme/poe_unoffical_api/internal/log"
)

type errorInfo struct {
	Time           string `json:"time"`
	Alarm          string `json:"alarm"`
	Message        string `json:"message"`
	Filename       string `json:"filename"`
	Line           int    `json:"line"`
	FuncName       string `json:"funcName"`
	DebugBacktrace string `json:"debug_backtrace"`
}

func DeferFunc(out ...interface{}) {
	if err := recover(); err != nil {
		log.Panic(err, out)
	}
}
