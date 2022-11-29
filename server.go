package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/LiZi-77/ActionSystem/proto"

	"google.golang.org/grpc"
)

type server struct {
	gRPC.UnimplementedAuctionServer        // You need this line if you have a server
	name                             string // Not required but useful if you want to name your server
	port                             string // Not required but useful if your server needs to know what port it's listening to
	highestPrice	 				 int32
	status 							 string // the status of current bid
	bidTimes						 int32	// total bid times, max 5
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "PrimaryReplica", "Server name") // set with "-name <name>" in terminal
var port = flag.String("port", "5001", "Server port")           // set with "-port <port>" in terminal default 5001

func main() {

	f := setLog() //uncomment this line to log to a log.txt file instead of the console
	defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// starts a goroutine executing the launchServer method.
	launchServer()

	// code here is unreachable because launchServer occupies the current thread.
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &server{
		name:           *serverName,
		port:           *port,
		highestPrice:   0,
		status:			"on going",
		bidTimes:       0,
	}

	gRPC.RegisterAuctionServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

//func bid
func (s *server) Bid(context context.Context, bid *gRPC.BidRequest) (*gRPC.Ack, error){
	log.Printf("Client %v bid amount %d", bid.ClientId, bid.Amount)

    if int(bid.Amount) > int(s.highestPrice) {
        s.highestPrice = int32(bid.Amount)
        //change the bid status
		s.bidTimes ++
		if s.bidTimes >= 5 {
			// when the total bid time > 5, this round is over
			s.status = "over"
			log.Printf("This bid round is over(bid times more than 5)")
			println("This bid round is over(bid times more than 5)")

			//start another round
			s.highestPrice = 0
			s.bidTimes = 0
			s.status = "on going"
			log.Printf("Starting a new bidding round, starting amount %d",s.highestPrice)
		}
        return &gRPC.Ack{Ack: gRPC.Acks_ACK_SUCCESS}, nil
    } else if int(bid.Amount) < int(s.highestPrice) {
        return &gRPC.Ack{Ack: gRPC.Acks_ACK_FAIL}, nil
    } else {
        return &gRPC.Ack{Ack: gRPC.Acks_ACK_EXCEPTION}, nil
    }
}


func (s *server) Result(context context.Context, empty *gRPC.Empty) (*gRPC.Outcome, error) {
	log.Printf("Current bid status is %s with the highest price %d",string(s.status),int32(s.highestPrice))
	println("Current bid status is %s with the highest price %d",string(s.status),int32(s.highestPrice))
	return &gRPC.Outcome{BidState: string(s.status), HighestPrice: int32(s.highestPrice)}, nil
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}