/home/hklett/sw/bin/protoc --go_out=plugins=grpc:. src/penbot/shared/*.proto

go build penbot/controller

export GOARCH=arm
go build penbot/robot
copy to robot and run
