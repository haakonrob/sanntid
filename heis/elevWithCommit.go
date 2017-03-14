package main

import(
	"fmt"
	"time"
    "./network"
    "./operator"
	PC"./commitprotocol"
	elev"./dummydriver"
)

var 
(
	globalActiveJobs []job
	online bool
	LocalID string
	activeIDs []string
)


func main(){

	newOrderCh := make(chan elev.Order)
	completedOrderCh := make(chan elev.Order)
	eventCh:= make(chan elev.Event)
	operatorCh := make(chan )

	incomingCh := make(chan interface{})
	outgoingCh := make(chan interface{})
	networkCh := make(chan network.Status)
	ticker := make(chan bool)

		
	var jobCount int
	tick := time.NewTicker(time.Millisecond * 250)

    go func() {
        for range tick.C {
        	if online {
            	ticker<- true
            }
        }
    }()
	
	for {
		select {
		case stat := <- networkCh:
			online, LocalID, activeIDs = stat.Online, stat.LocalID, stat.ActiveIDs
		case <-ticker:
			t, ok := PC.Transaction{}, true
			for ok {
				t, ok = PC.PopFromBuffer()
				// Add rejection and scroing logic here
				val := ProcessTransaction(t)
				t = PC.Sign(t, val)
				outgoingCh <- t
			}
		case newOrder := <- newOrderCh:
			if (!online) {
				operatorChan <- newOrder
				break
			}

			var success, accepted, committed bool
			if signatures, success := GetSignatures(newJob) ; success {
				newJob := Job {
					JobID:		jobCount,
					Owner:		mapMax(signatures),
					Timestamp:	time.Now(),
					Content:	newOrder,
				}
			}
			if success {
				accepted := QueryToCommit(newJob, PC.ADD, incomingCh, outgoingCh)
			}
			if accepted {
				committed := Commit(newJob, incomingCh, outgoingCh)
			}
			if committed {
				jobCount++
				globalActiveJobs = append(globalActiveJobs, newJob)
				operatorChan <- getLocalJobs(globalActiveJobs)
			}
			
		case doneOrder := <- completedCh:
			if (!online || doneOrder.OrderType == elev.COMMAND) {
				removeJob(doneOrder)
				break
			}

			var success, accepted, committed bool
			accepted = QueryToCommit(newJob, PC.REMOVE, incomingCh, outgoingCh)
			if accepted {
				committed := Commit(newJob, incomingCh, outgoingCh)
			}
			if committed {
				globalActiveJobs = append(globalActiveJobs, newJob)
			}

		case transaction := <- incomingCh:
			ProcessTransaction(transaction, Jobs)  //, conditions)
			outgoingCh <- transaction
		}
	}
}

func mapMax(m map[string]int)(string){
	max := 0
	maxKey := ""
	for key, val := range m { 
		if val > max {
			max = val
			maxKey = key
		}
	}
	return maxKey
}

func removeJob(doneJob job){
	for i, j := range(globalActiveJobs){
		if j.JobID == doneJob.JobID {
			globalActiveJobs = append(globalActiveJobs[:i], globalActiveJobs[i+1:]...)
		}
	}
}


func ProcessTransaction(t PC.Transaction, incomingCh, outgoingCh)(bool){
	data := (t.Content).(job)
	
	switch (t.Request) {
	case PC.SIGNATURE:
		t = t.Sign(t, operator.ScoreOrder(order))
		outgoingCh <- t
		if t, ok = readIncomingWithTimeout() ; ok {
			if (t.Request != PC.SIGNATURE) && (t.Request != PC.SIGNATURE){
				return ProcessTransaction(t, incomingCh, outgoingCh)
			}
			return false
		}

	case PC.ADD:
		for i, j := range(globalActiveJobs){
			if j.JobID == data.JobID {
				globalActiveJobs = append(globalActiveJobs[:i], globalActiveJobs[i+1:]...)
			}
		}
		t = t.Sign(t, 1)
		outgoingCh <- t
		if t, ok = readIncomingWithTimeout() ; ok {
			if (t.Request != PC.SIGNATURE) && (t.Request != PC.SIGNATURE){
				return ProcessTransaction(t, incomingCh, outgoingCh)
			}
			return
		}

	case PC.REMOVE:

	case PC.MOVE:
		fmt.Println("MOVE transaction not implemented.")
	case COMMIT:

	default:
		fmt.Println("Error, bad transaction query.")
		return 0
	}

	readIncomingWithTimeout := func(){
		select {
	    case res := <-incomingCh:
	        return res.(PC.Transaction), true
	    case <-time.After(time.Millisecond * 200):
	        return PC.Transaction{}, false
    }(PC.Transaction, bool)

	
	
	
}