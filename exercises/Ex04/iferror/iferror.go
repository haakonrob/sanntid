package iferror

import (
	"os"
)

type Action func()

func Quit(){
	os.Exit(0)
}

func Ignore(){
}
