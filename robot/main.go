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
}

func (s *Server) EnqueuePosition(
	ctx context.Context, request *pb.EnqueuePositionRequest) (*pb.EnqueuePositionResponse, error) {
	a1 := left_angle(request.P, s.Config)
	a2 := right_angle(request.P, s.Config)
	leftMotorPosition := to_motor_position(a1)
	rightMotorPosition := to_motor_position(a2)
	fmt.Println("Position: (%f, %f)", request.P.X, request.P.Y)
	fmt.Println("Left angle: %f", a1)
	fmt.Println("Right angle: %f", a2)
	fmt.Println("Left motor position: %d", leftMotorPosition)
	fmt.Println("Right motor position: %d", rightMotorPosition)
	s.Left.SetPosition(leftMotorPosition)
	s.Right.SetPosition(rightMotorPosition)
	s.Left.Go()
	s.Right.Go()
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

// x: -70mm min, 70mm max
// y: 70mm min, 120mm max
func main() {
	config := Config{
		A: 32, // 32mm
		L: 98, // 98mm
		R: 80, // 80mm
	}

	leftMotor := InitMotor("outA")
	rightMotor := InitMotor("outB")

	lis, err := net.Listen("tcp", ":4321")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPenBotServer(s, &Server{
		Config: &config,
		Left:   leftMotor,
		Right:  rightMotor,
	})
	reflection.Register(s)
	fmt.Println("serving")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
