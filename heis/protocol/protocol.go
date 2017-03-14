package protocol

var transactioncount int

type request int 
const (
	SCORE request = iota
	ADD
	REMOVE
	MOVE
)


type Transaction struct {
	SenderID string
	Request request
	Content interface{}
	Signatures map[string]int
}


func Sign(t Transaction, ID string, value int)(Transaction){
	t.Signatures[ID] = value
	return t
}

func GenerateTransaction(data interface{}, ID string, req request)(Transaction){
	// Signs once it's received the packet again
	t := Transaction{
		SenderID:			ID,
		Request:			req,
		Content:			data,
		Signatures: 		make(map[string]int),
	}	
	return t
}