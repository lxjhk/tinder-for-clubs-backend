package common

import "log"

func ErrFatalLog(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

