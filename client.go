package main

import (
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
	serverCount int
	currentBid  int
	roundOver   bool
	chanDone    []chan bool
	servers     []gRPC.AuctionClient
)

func main() {
	rand.Seed(time.Now().UnixNano())
	args := os.Args[1:]
	aid, _ := strconv.ParseInt(args[0], 10, 32)
	id = int32(aid)
	serverCount, _ := strconv.ParseInt(args[1], 10, 32)
	bidCount, _ := strconv.ParseInt(args[2], 10, 32)
	chanDone = make([]chan bool, int(serverCount))
	servers = make([]gRPC.AuctionClient, int(serverCount))

	for i := 0; i < int(serverCount); i++ {
		go ConnServer(i)
	}

	for i := 0; i < int(bidCount); i++ {
		log.Printf("Bidding round %d/%d", i+1, int(bidCount))
		Client(int(serverCount))
	}
}

// func CheckServer(err error, serverId int) bool {
// 	if err != nil {
// 		log.Printf("Server %d disconnected", serverId)
// 		servers[serverId] = nil
// 		chanDone[serverId] <- true
// 		return false
// 	}
// 	return true
// }

func BroadcastBid(amount int32) {
	lamport++
	timeout, _ := context.WithTimeout(context.Background(), time.Second*5)
	currBid := gRPC.BidRequest{ClientId: id, Amount: amount, Lamport: lamport}
	log.Printf("Client %v bidding %v", id, amount)
	for i, s := range servers {
		if s != nil {
			ack, err := s.Bid(timeout, &currBid)

			check := true
			if err != nil {
				log.Printf("Server %d disconnected", i)
				servers[i] = nil
				chanDone[i] <- true
				check = false
			}

			if check {
				if ack.Ack == gRPC.Acks_ACK_SUCCESS {
					log.Printf("Server %d accepted bid", i)
				} else if ack.Ack == gRPC.Acks_ACK_FAIL {
					log.Printf("Server %d rejected bid", i)
				} else if ack.Ack == gRPC.Acks_ACK_EXCEPTION {
					log.Printf("Server %d exception", i)
				}
			}
		}
	}
}

func GetResult() *gRPC.Outcome {
	lamport++
	timeout, _ := context.WithTimeout(context.Background(), time.Second*5)
	for i := 0; i < serverCount; i++ {
		if servers[i] != nil {
			result, err := servers[i].Result(timeout, &gRPC.Empty{})
			check := true
			if err != nil {
				log.Printf("Server %d disconnected", i)
				servers[i] = nil
				chanDone[i] <- true
				check = false
			}
			if !check {
				continue
			}
			return result
		}
	}
	return nil
}

// Client: Front end of the client
func Client(servers int) {
	for {
		if serverCount == servers {
			result := GetResult()

			if result.BidState != "on going" {
				if !roundOver {
					roundOver = true
					currentBid = 0
					log.Printf("Round over, bid state: %s, total bidding amount: %v", result.BidState, result.HighestPrice)
					return
				}
				time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))
				continue
			}

			roundOver = false
			currentBid += rand.Intn(50) + 1
			BroadcastBid(int32(currentBid) + result.HighestPrice)
			time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))
		}
	}
}

// DialServer: Dial server and connect to it
func ConnServer(serverId int) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", serverId+5000), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	servers[serverId] = gRPC.NewAuctionClient(conn)
	chanDone[serverId] = make(chan bool)
	log.Printf("Client %v bidding...", id)
	serverCount++

	<-chanDone[serverId]
	conn.Close()
}
