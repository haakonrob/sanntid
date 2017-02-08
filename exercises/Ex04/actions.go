package actions

import (
	"os"
)

type ActionFunc func()

func Quit(){
	os.Exit(0)
}

func Ignore(){
}
