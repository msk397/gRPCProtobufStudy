package main

import (
	"fmt"
	"gRPCProtobufStudy/service"
	"google.golang.org/protobuf/proto"
)

func main() {
	user := &service.User{
		Username: "msk",
		Age:      22,
	}
	//序列化
	marshal, err := proto.Marshal(user)
	if err != nil {
		panic(err)
	}

	//反序列化
	newUser := &service.User{}
	err = proto.Unmarshal(marshal, newUser)
	if err != nil {
		panic(err)
	}
	fmt.Println(newUser.Username)

}
