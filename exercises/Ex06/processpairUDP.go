package main


import(
	"os"
	"os/exec"
	//"syscall"
	"fmt"
	"time"
	"net"
)

const (
	port = ":20022"
)

var (
	currentCount int
)

func main(){
	currentCount = 1
	// If there is an argument -b, this starts as a backup process
	if len(os.Args) > 1 {
		if os.Args[1] == "-b" {
			startBackupWait()
		} else {
			fmt.Println("Wrong usage")
			os.Exit(0)
		}
	}
	
	backup := exec.Command("gnome-terminal", "-x","sh", "-c", "./processpair -b")
	backup.Run()
	startPrimary(currentCount, time.Now())
}

func startPrimary(count int, lastbcast time.Time){
	fmt.Println("Primary PID", os.Getpid())
	for {
		if time.Since(lastbcast) > time.Second {
			fmt.Println(count)
			count++
			lastbcast = time.Now()
			localhostBcast(count)
			
		} else {	
			time.Sleep(time.Second/8)
			localhostBcast(count)
		}
		
	}
}

func startBackupWait(){
	fmt.Println("Backup PID", os.Getpid() )
	
	ping := make(chan bool)
	addr, _ := net.ResolveUDPAddr("udp", port)
	sockln, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
	}
	
	listen := func(){
		buf := make([]byte, 1024)
		for {
			_, _, _ = sockln.ReadFromUDP(buf)
			//fmt.Println("Update",buf[0])	
			currentCount = int(buf[0])
			ping<-true	
		}
	}
	go listen()
	lastMsg := time.Now()
	for {
		time.Sleep(time.Millisecond)
		select {
			case msg := <-ping:
				if msg {
					lastMsg = time.Now()
				}
			default:
				if time.Since(lastMsg) > time.Second {
					fmt.Println("dead")
					sockln.Close()
					return
				}
			}				
	}
}

func listenUDP(output chan int, doneC chan bool){
	addr, _ := net.ResolveUDPAddr("udp", port)
	sockln, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
	}
	buf := make([]byte, 1024)
	for {
		select {
			case done := <-doneC:
				if done {
					fmt.Println("closing conn")
					defer sockln.Close() 
					return 
				}
			default:
				_, _, _ = sockln.ReadFromUDP(buf)
				//fmt.Println("Update",buf[0])	
				output<- int(buf[0])//int(buf[0:n])
		}
	}
}

func localhostBcast(count int){
	localhost, _ := net.ResolveUDPAddr("udp","localhost"+port)
	conn, _ := net.DialUDP("udp", nil, localhost)
	_, _ = conn.Write([]byte(string(count)))
	conn.Close()
}
