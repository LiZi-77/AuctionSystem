package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	gRPC "github.com/LiZi-77/ActionSystem/proto"
	"google.golang.org/grpc"
)

var (
	id          int32
	lamport     int64
	currentBid  int
	chanDone    []chan bool
	servers     []gRPC.AuctionClient // all possible servers
)

func CheckServer(err error, serverId int) bool {
    if err != nil {
        log.Printf("Server %v unresponsive, connection disconnected", serverId)
        servers[serverId] = nil
        //chanDone[serverId] <- true
        return false
    }
    return true
}

func FrontEndResult(clientId int) *gRPC.Outcome {
	//println("client %d send the Result Request", clientId)
	log.Printf("client %d send the Result Request", clientId)
	lamport++
	timeout, _ := context.WithTimeout(context.Background(), time.Second*5)
    for i := 0; i < 2; i++ {
		//println("send to server ", i)
        if servers[i] != nil {
			//println("server ", i, "exists")
	        result, err := servers[i].Result(timeout, &gRPC.Empty{})
			//println(err)
			//println(CheckServer(err,i))
	        if !CheckServer(err,i) {
				//println("about to continue to another server")
                continue
            }
			//println("about to return a result")
            return result
        }
		//println("server ", i, "doesn't exist")
    }
    return nil
}

func FrontEndBid(clientId int, amount int) {
	log.Printf("client %d send the Bid Request", clientId)
	lamport++
	timeout, _ := context.WithTimeout(context.Background(), time.Second*5)
	for i := 0; i < 2; i++ {
        if servers[i] != nil {
	        ack, err := servers[i].Bid(timeout, &gRPC.BidRequest{ClientId: int32(clientId), Amount: int32(amount), Lamport: lamport})
	        if CheckServer(err,i) {
                switch ack.Ack {
				case gRPC.Acks_ACK_FAIL:
					log.Printf("Client %v bid %v on server %v failed!",clientId, amount, i)
					println("Failed! Your bit is less than other users or the auction is over now.")
				case gRPC.Acks_ACK_SUCCESS:
					log.Printf("Client %s bid %v on server %v failed!",clientId, amount, i)
					println("Sucess! Your bit is higher than other users.")
				case gRPC.Acks_ACK_EXCEPTION:
					println("Exception!")
					log.Printf("Bidding server %v exception",i)
				}
            }
        }
    }
}


func bidStart(clientId int) {
	println("Bid already start now.")

	scanner := bufio.NewScanner(os.Stdin)
	println("Welcome to the Auction System!\n 1.Result - to check current highest price \n 2. Bid - to give your amount")

	for {
		scanner.Scan()
		text := scanner.Text()

		if text == "1" {
			// Result method
			result := FrontEndResult(clientId)
			if(result.BidState == true){
				println("The auction is still open, current hightest bid is",result.HighestPrice)
			} else {
				println("The auction is over, the hightest bid is",result.HighestPrice)
			}
		} else if text == "2" {
			println("please input the amount:")
			// Bid method
			price := 0
			_, _ = fmt.Scanln(&price)

			if(price > currentBid) {
				currentBid = price
				//result := FrontEndBid(clientId, price)
				FrontEndBid(clientId, price)
				// if(result.Ack == gRPC.Acks_ACK_SUCCESS){
				// 	println("Success: your bid is higher than anyone else!")
				// } else if (result.Ack == gRPC.Acks_ACK_FAIL){
				// 	println("Fail: your bid is less than some one else or the auction is over.")
				// } else {
				// 	println("Exception: some error occurs.")
				// }
			}
			// } else {
			// 	println("Your last bid is ", currentBid, ", please give a higher bid." )
			// }
		} else {
			println("Please input 1 or 2")
		}
	}
}
func connServer(serverId int, clientId int){
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", serverId+5001), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect server: %v", err)
	}

	servers[serverId] = gRPC.NewAuctionClient(conn)
    chanDone[serverId] = make(chan bool)
	log.Printf("Client %d connected to server %d",clientId, serverId)

	//<-chanDone[serverId]
    //conn.Close()
}

func main() {
	f := setLog() //uncomment this line to log to a log.txt file instead of the console
	defer f.Close()

	rand.Seed(time.Now().UnixNano())
	args := os.Args[1:]

	// go run client.go <clientNo.>
	clientNo, _ := strconv.ParseInt(args[0], 10, 32)
	//there are 2 severs at port 5001 and 5002

	id = int32(clientNo)
	chanDone = make([]chan bool, 2)
	servers = make([]gRPC.AuctionClient, 2)	// store all servers

	for i := 0; i < 2; i++ {
		connServer(i, int(clientNo))
	}

	//println(GetResult().HighestPrice)
	bidStart(int(clientNo))
}

func setLog() *os.File {

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
