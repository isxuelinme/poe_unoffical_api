package core

import "github.com/isxuelinme/poe_unoffical_api/internal/log"

func init() {
	log.SetLogMode(log.DEBUG)
}

var (
	LOG_DEBUG   log.Mode = log.DEBUG
	LOG_SILENCE log.Mode = log.SILENCE
	LOG_ERROR   log.Mode = log.ERROR
)

func SetLogMode(mode log.Mode) {
	log.SetLogMode(mode)
}
