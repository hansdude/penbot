package main

import (
	"fmt"
	"log"
	"math"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "penbot/shared"
)

const (
	scale = 1
  xDim = 1000
  yDim = 1000
)

func main() {
	gtk.Init(nil)

	conn, err := grpc.Dial("ev3dev.local:4321", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPenBotClient(conn)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatalf("Unable to create window: %v", err)
	}
	win.SetTitle("Robot Controller")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.Connect("motion-notify-event", func(win *gtk.Window, evt *gdk.Event) {
    x, y := gdk.EventMotionNewFromEvent(evt).MotionVal()
    xLimit := math.Max(0, math.Min(x, xDim * scale))
    yLimit := math.Max(0, math.Min(y, yDim * scale))
    fmt.Printf("%f, %f\n", xLimit, yLimit)
    requestPoint := pb.Point{
      X: (xLimit / scale) - (xDim / 2),
      Y: (yDim / 2) - (yLimit / scale) + yDim / 2, // 70mm offset
    }
    fmt.Println(requestPoint)
    _, err = c.EnqueuePosition(context.Background(), &pb.EnqueuePositionRequest{P: &requestPoint})
    if err != nil {
      log.Fatalf("could not enqueue position: %v", err)
    }
	})

	l, err := gtk.LabelNew("Click here and drag to move.")
	if err != nil {
		log.Fatalf("Unable to create label: %v", err)
	}

	win.Add(l)

	win.SetDefaultSize(xDim*scale, yDim*scale)

	win.ShowAll()

	gtk.Main()
}
