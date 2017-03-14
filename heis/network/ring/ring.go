package ring

import (
	//"../conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"
)

func test() {
    chtx := make(chan string)
    chrx := make(chan string)
    updateCh := make(chan string)
    go Receiver(20000, chrx)
    go Transmitter(updateCh, chtx)
    updateCh<-"me"
    for {
    }
}

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func Transmitter(targetCh chan string, chans ...interface{}) {
	checkArgs(chans...)

	n := 0
	for _, _ = range chans {
		n++
	}
    
	selectCases := make([]reflect.SelectCase, n)
	typeNames := make([]string, n)
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

    var initialised bool
	var conn net.Conn
	var errorCount int

	go func() {
	    var addr string
		for {
			if initialised {
			    addr = <-targetCh
			    initialised = false
			    conn.Close()
			}
			c, err := net.Dial("tcp", addr)
			if err == nil {
			    initialised = true
			    conn = c
			} 
			    
		}
	}()

	for {
		for !initialised{ time.Sleep(time.Millisecond*100) }
		chosen, value, _ := reflect.Select(selectCases)
		buf, _ := json.Marshal(value.Interface())
		_, err := conn.Write([]byte(typeNames[chosen]+string(buf)))
		if err != nil {
		    errorCount++
		    time.Sleep(time.Millisecond*100)
		    if errorCount > 5 {
		        initialised = false
		    }
		}
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
    
    var initialised bool
	var buf [1024]byte
	
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
        if err != nil {
            fmt.Println("Unable to listen on port ", port)
        }
	var conn net.Conn
	for {
	    switch initialised {
	    case false:
	        c, err := ln.Accept()
        	if err == nil {
        		initialised = true
        		conn = c
        	}
        	
        case true:
    		n, err := conn.Read(buf[0:])
    		if err != nil {
    		    initialised = false
    		} else {    
        		for _, ch := range chans {
        			T := reflect.TypeOf(ch).Elem()
        			typeName := T.String()
        			if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
        				v := reflect.New(T)
        				json.Unmarshal(buf[len(typeName):n], v.Interface())
        				reflect.Select([]reflect.SelectCase{{
        					Dir:  reflect.SelectSend,
        					Chan: reflect.ValueOf(ch),
        					Send: reflect.Indirect(v),
        				}})
        			}
        		}
    		}
	    }
	}
}


// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for _, _ = range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg#%d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		switch elemType.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
				elemType.String(), i+1))
		case reflect.Map:
			if elemType.Key().Kind() != reflect.String {
				panic(fmt.Sprintf(
					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
					elemType.String(), i+1))
			}
		}
	}
}
