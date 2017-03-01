package main 

import (
    "time"
    "fmt"
    _"io"
    "io/ioutil"
    "os"
)

func main(){
	c := make([]byte,1)
	
	for{
		atime := time.Now()
		mtime := time.Now()

		if c[0] < 5{
				c[0] = 0
				fmt.Println(c[0])
				//send to backup
				err := ioutil.WriteFile("backup", c, os.ModeDevice)
				CheckError(err)
				err1 := os.Chtimes("backup", atime, mtime)
				CheckError(err1)


		} else{
				c[0] = c[0] + 1
				fmt.Println(c[0])
				//send to backup
				err := ioutil.WriteFile("backup", c, os.ModeDevice)
				CheckError(err)
				err1 := os.Chtimes("backup", atime, mtime)
				CheckError(err1)

			}
		}

}


func CheckError(e error){
	if e!= nil{
		panic(e)	
	}
}