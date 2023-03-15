package main

import (
	"github.com/isxuelinme/poe_unoffical_api/client/cli"
	"github.com/isxuelinme/poe_unoffical_api/client/sse"
	"github.com/isxuelinme/poe_unoffical_api/core"

	"os"
)

func main() {
	core.SetLogMode(core.LOG_ERROR)
	if os.Getenv("RUN_MODE") == "SSE" {
		sse.SSE()
	} else {
		cli.CLI()
	}
	select {}

}
