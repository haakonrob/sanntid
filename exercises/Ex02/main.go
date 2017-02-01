// Go 1.2
// go run helloworld_go.go

package main

import (
    "fmt"
    "runtime"
    
)


var x int = 0

func f1(ch, ch1 chan int) {

    for i:=0;i < 1000000; i++{
        <- ch
        x = x+1
        ch <- 1
    }
    ch1 <- x
}

func f2(ch, ch1 chan int) {
    for j:=0;j < 1000000; j++{
        <- ch
        x = x-1
        ch <-1
    }
    ch1 <- x

}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())    // I guess this is a hint to what GOMAXPROCS does...
                                            // Try doing the exercise both with and without it!
    

    ch := make(chan int, 1000000)
    ch1 := make(chan int, 1)
    ch <- 1

    go f1(ch, ch1)                      // This spawns someGoroutine() as a goroutine
    go f2(ch, ch1)

    <-ch1

    // We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
    // We'll come back to using channels in Exercise 2. For now: Sleep.
    Println("Go", <-ch1)
}
