This is EX04.

We will use TCP for comms, and UDP croadcast to initialise the network of elevators.
No master/slave, only identical processes.
For the 3 elevator problem, all elevators can connect to each other. For N elevs, they can connect in a ring, with ascending IP addresses/IDs denoting their order.


Network module specs:
Sequence
Init: 
	"Spam" broadcast with own IP and passcode.
	Identify other elevs through UDP broadcast. Save network state.
	Establish TCP "ring".
	Start sending the packet.
main:
	receive packet
	send packet to coordinator
	receive packet from coordinator. Coordinator tasks:
		Update CLK variable for every send.
		If a new local order has been registered, set it active, and if successful send confirmation to coordinator (so it can turn on the light).
		Score all active orders.
		If order is fully scored, allocate order.
		If allocated order is completed, set order to inactive.
	send incremented packet to next elev.
	Send incremental package continually around ring. 
		
		
		
			
	//opt: print packet data.
errors:
	Next neighbour falls out of network: Close socket and open socket to next neighbour, deactivate the order and rescore.
	Alone in network: Close all sockets, continue UDPlisten, send nada.
	An elevator order timeouts: Allocate order to another elev.
