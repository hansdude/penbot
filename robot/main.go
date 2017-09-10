package main

import (
	"fmt"
	"math"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "penbot/shared"
)

const (
	Left  = 1
	Right = -1
)

type Config struct {
	A float64
	L float64
	R float64
}

func find_angle(p *pb.Point, l float64, r float64, leftOrRight float64) float64 {
	v := p.X*p.X + p.Y*p.Y
	u := l*l - v + r*r
	t := p.X * math.Sqrt(4*l*l*r*r-u*u)
	y := (leftOrRight*t + p.Y*(v+r*r-l*l)) / (2 * v)
	return math.Asin(y / r)
}

func left_angle(p *pb.Point, config *Config) float64 {
	return find_angle(&pb.Point{X: p.X + config.A, Y: p.Y}, config.L, config.R, Left)
}

func right_angle(p *pb.Point, config *Config) float64 {
	return find_angle(&pb.Point{X: p.X - config.A, Y: p.Y}, config.L, config.R, Right)
}

type Server struct {
  Config *Config
}

func (s *Server) EnqueuePosition(
	ctx context.Context, request *pb.EnqueuePositionRequest) (*pb.EnqueuePositionResponse, error) {
	fmt.Println("Position: (%f, %f)", request.P.X, request.P.Y)
	a1 := left_angle(request.P, s.Config)
	a2 := right_angle(request.P, s.Config)
	fmt.Println(a1)
	fmt.Println(a2)
	return &pb.EnqueuePositionResponse{}, nil
}

func main() {
	config := Config{
		A: 32, // 32mm
		L: 98, // 98mm
		R: 80, // 80mm
	}

	lis, err := net.Listen("tcp", ":4321")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPenBotServer(s, &Server{Config:&config})
	reflection.Register(s)
	fmt.Println("serving")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// TODO:
// turn off the brake and get the angle limits
// 3x gear ratio
// -195 lower limit 40 upper limit
// x: -70mm min, 70mm max
// y: 70mm min, 120mm max
