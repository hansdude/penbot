package main

import (
	"fmt"
	"log"
	"math"
	"net"

	"github.com/ev3go/ev3dev"
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

func to_motor_position(angle float64) int {
	// Go from radians to degrees.
	scaled := (angle / math.Pi) * 180.0
	//TODO: maybe change this to be an initialization step
	// Rotate by -90*
	shifted := (scaled - 90.0)
	// Compensate for the 3:1 gear ratio.
	gearScaled := shifted * 3.0
	// Enforce -195 lower limit 40 upper limit.
	if gearScaled < -195.0 {
		log.Printf("Position %f was below lower limit -195", gearScaled)
		return -195
	}
	if gearScaled > 40.0 {
		log.Printf("Position %f was above upper limit 40", gearScaled)
		return 40
	}
	return int(math.Floor(gearScaled + .5))
}

type Server struct {
	Config *Config
	Left   *Motor
	Right  *Motor
	Commands chan *pb.Point
}

func (server *Server) EnqueuePosition(
	ctx context.Context, request *pb.EnqueuePositionRequest) (*pb.EnqueuePositionResponse, error) {
	server.Commands <- request.P
	return &pb.EnqueuePositionResponse{}, nil
}

type Motor struct {
	Ev3Motor *ev3dev.TachoMotor
}

func InitMotor(output string) *Motor {
	motor, err := ev3dev.TachoMotorFor(output, "lego-ev3-l-motor")
	if err != nil {
		log.Fatalf("count not get motor %s: %v", output, err)
	}
	err = motor.SetStopAction("brake").Err()
	if err != nil {
		log.Fatalf("could not set stop action: %v", err)
	}
	motor.SetSpeedSetpoint(100)
	return &Motor{Ev3Motor: motor}
}

func (motor *Motor) SetPosition(position int) {
	motor.Ev3Motor.SetPositionSetpoint(position)
}

func (motor *Motor) Go() {
	motor.Ev3Motor.Command("run-to-abs-pos")
}

func (server *Server) Run() {
	for point := range server.Commands {
		a1 := left_angle(point, server.Config)
		a2 := right_angle(point, server.Config)
		leftMotorPosition := to_motor_position(a1)
		rightMotorPosition := to_motor_position(a2)
		fmt.Printf("Position: (%f, %f)", point.X, point.Y)
		fmt.Printf("Left angle: %f", a1)
		fmt.Printf("Right angle: %f", a2)
		fmt.Printf("Left motor position: %d", leftMotorPosition)
		fmt.Printf("Right motor position: %d", rightMotorPosition)
		server.Left.SetPosition(leftMotorPosition)
		server.Right.SetPosition(rightMotorPosition)
		server.Left.Go()
		server.Right.Go()
	}
}

// x: -70mm min, 70mm max, 140mm total
// y: 70mm min, 120mm max, 50mm total
func main() {
	config := Config{
		A: 32, // 32mm
		L: 98, // 98mm
		R: 80, // 80mm
	}

	//TODO: garbage collection?
	leftMotor := InitMotor("outA")
	rightMotor := InitMotor("outB")

	robotServer := Server{
		Config: &config,
		Left:   leftMotor,
		Right:  rightMotor,
		Commands: make(chan *pb.Point),
	}

	go robotServer.Run()

	lis, err := net.Listen("tcp", ":4321")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPenBotServer(s, &robotServer)
	reflection.Register(s)
	fmt.Println("serving")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
