package fmutil

import (
	"fmt"
	"unicode"
)

func MustNoError(err error, msgArgs ...any) {
	if err != nil {
		if len(msgArgs) == 0 {
			panic(err)
		} else {
			args := msgArgs[1:]
			var msg string
			switch v := msgArgs[0].(type) {
			case string:
				msg = v
			case fmt.Stringer:
				msg = v.String()
			default:
				msg = fmt.Sprintf("%v", msg)
			}
			runes := []rune(msg)
			if len(runes) > 0 {
				lastChar := runes[len(runes)-1]
				if unicode.IsLetter(lastChar) {
					msg += "."
				}
			}
			err = fmt.Errorf(msg+" err:%w", append(args, err)...)
			panic(err)
		}
	}
}

func MustNoError2(_ any, err error, msgArgs ...any) {
	MustNoError(err, msgArgs...)
}

func MustTrue(v bool, msg string, args ...any) {
	if !v {
		panic(fmt.Errorf(msg, args...))
	}
}

func Panic(msg string, args ...any) {
	panic(fmt.Errorf(msg, args...))
}
