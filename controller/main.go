package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "penbot/shared"
)

func main() {
  conn, err := grpc.Dial("ev3dev.local:4321", grpc.WithInsecure())
  if err != nil {
    log.Fatalf("could not connect: %v", err)
  }
  defer conn.Close()
  c := pb.NewPenBotClient(conn)

  _, err = c.EnqueuePosition(context.Background(), &pb.EnqueuePositionRequest{P:&pb.Point{X: 20, Y: 90}})
  if err != nil {
    log.Fatalf("could not enqueue position: %v", err)
  }
}

