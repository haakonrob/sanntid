package main

import (
    "bufio"
    "fmt"
    "io"
    "io/ioutil"
    "os"
)

func  main() {


	backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run backup.go")

	CreateFile()
	PrimaryProcess()
	for{
		if !IsAlive(){
			backup.Run()
		}
	}


	dat, err := ioutil.ReadFile("backup")
    CheckError(err)
    fmt.Print(dat)
    c := make([]byte, 1)
    c[0] = 1
		for{
			time.se
			c[0] = 0
			fmt.Println(c[0])
			//send to backup
			err := ioutil.WriteFile("backup", c, os.ModeDevice)
			CheckError(err)
			err1 := os.Chtimes("backup", atime, mtime)
			CheckError(err1)

}


func CreateFile() {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		CheckError(err)
		defer file.Close()
	}else{
		//readFile
	}
}