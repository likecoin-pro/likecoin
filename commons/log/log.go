package log

import (
	"io/ioutil"
	"log"
	"os"
)

const (
	LevelNone    = 0
	LevelFatal   = 1
	LevelError   = 2
	LevelWarning = 3
	LevelInfo    = 4
	LevelDebug   = 5
	LevelTrace   = 6
)

var (
	Fatal   = newLogger("⛔ FATAL: ")
	Error   = newLogger("🛑 ERROR: ")
	Warning = newLogger("⚠️️ WARNG: ")
	Info    = newLogger("")
	Debug   = newLogger("ℹ️ Debug: ")
	Trace   = newLogger("-- Trace: ")
)

func init() {
	SetLogLevel(LevelInfo)
}

func newLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, log.LstdFlags)
}

func turnOn(logger *log.Logger, on bool) {
	if on {
		logger.SetOutput(os.Stderr)
	} else {
		logger.SetOutput(ioutil.Discard)
	}
}

func SetLogLevel(lv int) {
	turnOn(Fatal, lv >= LevelFatal)
	turnOn(Error, lv >= LevelError)
	turnOn(Warning, lv >= LevelWarning)
	turnOn(Info, lv >= LevelInfo)
	turnOn(Debug, lv >= LevelDebug)
	turnOn(Trace, lv >= LevelTrace)
}

func Panic(v ...interface{}) {
	Fatal.Panic(v...)
}

func Print(v ...interface{}) {
	Info.Print(v...)
}

func Printf(format string, v ...interface{}) {
	Info.Printf(format, v...)
}
