package helpers

import (
	"log"
	"runtime"
)

func LogFatal(err error) {
	_, filename, line, _ := runtime.Caller(1)
	log.Fatalf("[error] in %s:%d\n%v", filename, line, err)
}
