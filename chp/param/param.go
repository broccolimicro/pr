package param

import (
	"strconv"
	"os"
	"strings"
	"runtime"
)

func Out() string {
	cwd := String(1, "")
	pc := make([]uintptr, 20)
	callers := runtime.Callers(0, pc)
	for i := 0; i < callers; i++ {
		name := runtime.FuncForPC(pc[i]).Name()
		start := strings.Index(name, "TestUnit")
		if start >= 0 {
			result := "test/chp/" + name[start:]
			if len(cwd) == 0 {
				return result
			}
			return cwd + "/" + result
		}
	}

	return "test/chp/Unknown"
}

func Profile() string {
	return os.Getenv("ACT_PROFILE")
}

func String(pos int, defaultValue string) string {
	i := len(os.Args)-1
	for ; i >= 0; i-- {
		if strings.HasPrefix(os.Args[i], "-test") {
			break
		}
	}
	if i + pos < len(os.Args) {
		return os.Args[i+pos]
	}
	return defaultValue
}

func Int64(pos int, defaultValue int64) int64 {
	i := len(os.Args)-1
	for ; i >= 0; i-- {
		if strings.HasPrefix(os.Args[i], "-test") {
			break
		}
	}
	if i + pos < len(os.Args) {
		value, err := strconv.ParseInt(os.Args[i+pos], 10, 64)
		if err == nil {
			return value
		}
	}
	return defaultValue
}

func Int(pos int, defaultValue int) int {
	i := len(os.Args)-1
	for ; i >= 0; i-- {
		if strings.HasPrefix(os.Args[i], "-test") {
			break
		}
	}
	if i + pos < len(os.Args) {
		value, err := strconv.ParseInt(os.Args[i+pos], 10, 64)
		if err == nil {
			return int(value)
		}
	}
	return int(defaultValue)
}


func Bool(pos int, defaultValue bool) bool {
	i := len(os.Args)-1
	for ; i >= 0; i-- {
		if strings.HasPrefix(os.Args[i], "-test") {
			break
		}
	}
	if i + pos < len(os.Args) {
		value, err := strconv.ParseBool(os.Args[i+pos])
		if err == nil {
			return value
		}
	}
	return defaultValue
}
