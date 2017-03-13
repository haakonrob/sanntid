package ringtcp

import (
	//"../conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
)

func main() {
    chtx := make(chan string)
    chrx := make(chan string)
    updateCh := make(chan string)
    go Receiver(20000, updateCh, chrx)
    go Transmitter(20000, , chtx)
    updateCh<-"me"
    for {
        dsafasdfasd    
    }
}

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func Transmitter(port int, targetCh chan string, chans ...interface{}) {
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

	go func() {
		for {
			if initialised {
			    addr = <-targetCh
			    intialised = false
			    conn.Close()
			}
			conn, err := net.Dial("tcp", addr)
			if err == nil {
			    initialised = true
			} 
			    
		}
	}()

	for {
	    if !initialised{  }
		chosen, value, _ := reflect.Select(selectCases)
		buf, _ := json.Marshal(value.Interface())
		conn.Write([]byte(typeNames[chosen]+string(buf)))
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
    
    var initialised bool
    reset := make(chan bool)
	var buf [1024]byte
	
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
        if err != nil {
            fmt.Println("Unable to listen on port ", port)
        }
	
	for {
	    switch initialised {
	    case false:
	        conn, err := ln.Accept()
        	if err == nil {
        		initialised = true
        	}
        	
        case true:
    		n, err := conn.Read(buf[0:])
    		if err != nil {
    		    initialised = false
    		} else {    
        		for _, ch := range chans {
        			T := reflect.TypeOf(ch).Elem()
        			typeName := T.String()
        		
        			if strings.HasPrefix(msg+"{", typeName) {
        				v := reflect.New(T)
        				json.Unmarshal([]byte(msg[1])[len(typeName):], v.Interface())
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

