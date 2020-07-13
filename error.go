package phppack

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

var (
	errNea = errors.New("not enough args")
)

func errT() error {
	return errors.New("type " + callerType() + ": wrong data type")
}

func errL() error {
	return errors.New("type " + callerType() + ": not enough characters in string")
}

func callerType() string {
	for i := 2; i < 6; i++ {
		pc, _, _, _ := runtime.Caller(i)
		n := runtime.FuncForPC(pc).Name()
		ts := strings.Split(n, "2")
		fmt.Println(n, ts)
		if len(ts) == 2 && len(ts[1]) < 3 {
			return ts[1]
		}
	}
	return ""
}
