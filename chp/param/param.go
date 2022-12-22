package param

import (
	"strconv"
	"os"
	"strings"
)

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
