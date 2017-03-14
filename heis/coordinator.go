package main

import(
	"fmt"
	"time"
    "./network"
    "./operator"
	PC"./protocol"
	elev"./dummydriver"
	"runtime"
)

type job struct {
	JobID 		int
	Owner 		string
	Timestamp	time.Time
	Order		elev.Order
}

var 
(
	globalActiveJobs []job
	jobCount int = 1
	online bool
	LocalID string
	activeIDs []string
)


func main(){
	//elev.Init()
	runtime.GOMAXPROCS(20)
	
	newOrderCh := make(chan elev.Order)
	completedOrderCh := make(chan elev.Order)
	eventCh:= make(chan elev.Event)
	operatorCh := make(chan elev.Order)

	incomingCh := make(chan interface{})
	outgoingCh := make(chan interface{})
	networkCh := make(chan network.Status)
		
	go network.Monitor(networkCh, true, "localhost", incomingCh, outgoingCh)
	go elevoperator.Start(eventCh, operatorCh, completedOrderCh)
	go elev.Poller(newOrderCh, eventCh)

	
	for {
		select {
		case stat := <- networkCh:
			online, LocalID, activeIDs = stat.Online, stat.LocalID, stat.ActiveIDs
		
		case newOrder := <- newOrderCh:
			if (!online || newOrder.Type == elev.BUTTON_COMMAND) {
				elev.SetButtonLamp(newOrder.Type, newOrder.Floor, true)
				operatorCh <- newOrder
				break
			}
				
			newJob := job {
				JobID:		-1,
				Owner:		"",
				Timestamp:	time.Now(),
				Order:		newOrder,
			}
			t := PC.GenerateTransaction(newJob, LocalID, PC.SCORE)
			t = PC.Sign(t, LocalID, elevoperator.ScoreOrder(newOrder))
			outgoingCh <- t
			
		case doneOrder := <- completedOrderCh:
			completedJobs := findCompletedJobs(doneOrder)
			
			for _, j := range(completedJobs) {
				if online {
					t := PC.GenerateTransaction(j, LocalID, PC.REMOVE)
					t = PC.Sign(t, LocalID, 1)
					outgoingCh <- t
				}
				removeJob(j)
			}

		case msg := <- incomingCh:
			ProcessTransaction(msg.(PC.Transaction), outgoingCh)  //, conditions)
		}
	}
}


func ProcessTransaction(t PC.Transaction, outgoingCh chan interface{})(){
	data := (t.Content).(job)
	
	switch (t.Request) {
	case PC.SCORE:
		if t.SenderID == LocalID {

			newJob := job {
				JobID:		jobCount,
				Owner:		mapMax(t.Signatures),
				Timestamp:	time.Now(),
				Order:		data.Order,
			}

			t = PC.GenerateTransaction(newJob, LocalID, PC.ADD)
			t = PC.Sign(t, LocalID, jobCount)
			outgoingCh <- t
		} else {
			t = PC.Sign(t, LocalID, elevoperator.ScoreOrder(data.Order))
			outgoingCh <- t
		}

	case PC.ADD:
		if addJob(data) {
			t = PC.Sign(t, LocalID, 1)
			if t.SenderID != LocalID {
				outgoingCh <- t
			}
		}

	case PC.REMOVE:
		if t.SenderID != LocalID {
			removeJob(data)
			t = PC.Sign(t, LocalID, 1)
			outgoingCh <- t
		}

	case PC.MOVE:
		if t.SenderID == LocalID {

			newJob := job {
				JobID:		jobCount,
				Owner:		mapMax(t.Signatures),
				Timestamp:	time.Now(),
				Order:		data.Order,
			}

			t_remove := PC.GenerateTransaction(newJob, LocalID, PC.REMOVE)
			t_add := PC.GenerateTransaction(newJob, LocalID, PC.ADD)
			t_remove = PC.Sign(t, LocalID, 1)
			t_add = PC.Sign(t, LocalID, 1)
			outgoingCh <- t_remove
			outgoingCh <- t_add

		} else {
			t = PC.Sign(t, LocalID, elevoperator.ScoreOrder(data.Order))
			outgoingCh <- t
		}
		
	default:
		fmt.Println("Error, bad transaction query.") 
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

func addJob(newJob job)(bool){
	if _, alreadyOrdered := findJob(newJob) ; !alreadyOrdered {
		globalActiveJobs = append(globalActiveJobs, newJob)
		elev.SetButtonLamp(newJob.Order.Type, newJob.Order.Floor , true)
		jobCount++
		return true
	}
	return false

}

func findJob(lostJob job) (int, bool) {
	for i, j := range(globalActiveJobs){
		if j.JobID == lostJob.JobID && (j.Owner == lostJob.Owner) {
			return i, true
		}
	}
	return 0, false
}

func removeJob(doneJob job)(bool){
	if i, inList := findJob(doneJob) ; inList {
		globalActiveJobs = append(globalActiveJobs[:i], globalActiveJobs[i+1:]...)
			elev.SetButtonLamp(doneJob.Order.Type, doneJob.Order.Floor , false)
			elev.SetButtonLamp(elev.BUTTON_COMMAND, doneJob.Order.Floor, false)
			return true
	}
	return false
}

func findCompletedJobs(completedOrder elev.Order)([]job){
	var completedJobs []job
	for _, job := range(globalActiveJobs){
		if job.Order.Floor == completedOrder.Floor {
			completedJobs = append(completedJobs, job)
		}
	}
	return completedJobs	
}