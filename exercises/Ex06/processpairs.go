package main

import (
	"fmt"
	"time"
	"io/ioutil"
	"os"
	"os/exec"
	)


//var path = "~/go/src/github.com/gitders222/sanntid/exercises/Ex06/backup.txt"
var path = "/backup.txt"
var count int = 0

func main(){

	backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run backup.go")


	//CreateFile()
	go PrimaryProcess()

	for{
		if !IsAlive(){
			fmt.Println("DEAD")
			backup.Run()
		}
	}

}



func PrimaryProcess(){
	c := make([]byte,1)
	
	for{
		atime := time.Now()
		mtime := time.Now()
		time.Sleep(time.Second * 2)
		
		c[0]++
		fmt.Println(c[0])
		//send to backup
		err := ioutil.WriteFile("backup.txt", c, os.ModeDevice)
		CheckError(err)
		err1 := os.Chtimes("backup.txt", atime, mtime)
		CheckError(err1)

	}
}


func CreateFile() {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		CheckError(err)
		defer file.Close()
	}
}

func IsAlive()bool{
	
	time.Sleep(time.Second * 2)
	dat, err := ioutil.ReadFile("backup.txt")
	fi, err := os.Stat("backup.txt")
	CheckError(err)
	_ = dat
	
	time := time.Now()
	timeFile := fi.ModTime
	fmt.Println(timeFile)
	fmt.Println(time)

	if timeFile > time{
		return true 
	}else{
		return false
	}
}

func Exit(file *os.File){
	defer os.Remove(file.Name())

}

func CheckError(e error){
	if e!= nil{
		panic(e)
		
	}
}