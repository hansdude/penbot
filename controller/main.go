package main

import (
	"flag"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "penbot/shared"
)

var x float64
var y float64

func main() {
	flag.Float64Var(&x, "x", 20.0, "The x coordinate to send to the server.")
	flag.Float64Var(&y, "y", 90.0, "The y coordinate to send to the server.")
	flag.Parse()

	requestPoint := pb.Point{X: x, Y: y}

	conn, err := grpc.Dial("ev3dev.local:4321", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPenBotClient(conn)

	_, err = c.EnqueuePosition(context.Background(), &pb.EnqueuePositionRequest{P: &requestPoint})
	if err != nil {
		log.Fatalf("could not enqueue position: %v", err)
	}
}
