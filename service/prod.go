package service

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"log"
	"sync"
	"time"
)

var ProdService = &prodService{}
var wg sync.WaitGroup

type prodService struct {
}

func (p *prodService) GetProductStock(context context.Context, request *ProductRequest) (*ProductResponse, error) {

	//实现具体业务逻辑
	stock := p.GetStockById(request.ProdId)
	user := User{
		Username: "msk",
	}
	content := Content{Msg: "123"}
	any, err := anypb.New(&content)
	if err != nil {
		return nil, err
	}
	return &ProductResponse{ProdStock: stock, User: &user, Data: any}, nil
}

func (p *prodService) UpdateStockClientStream(stream ProdService_UpdateStockClientStreamServer) error {
	count := 0
	for {
		//源源不断的接受客户端的信息
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		fmt.Println("服务端接受到的流：", recv.ProdId, count)
		count++
		if count > 10 {
			rep := &ProductResponse{ProdStock: recv.ProdId}
			err := stream.SendAndClose(rep)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func (p *prodService) GetProductStockServerStream(request *ProductRequest, stream ProdService_GetProductStockServerStreamServer) error {
	count := 10
	for {
		time.Sleep(time.Second)

		//返回nil,表示发送完成
		if count < 0 {
			content := Content{
				Msg: "已经结束发送，3秒钟后关闭",
			}
			a, err := anypb.New(&content)
			if err != nil {
				return err
			}
			rep := &ProductResponse{Data: a}

			err = stream.Send(rep)
			if err != nil {
				log.Fatalln(err)
			}
			time.Sleep(3 * time.Second)
			content = Content{
				Msg: "关闭连接",
			}
			a, err = anypb.New(&content)
			if err != nil {
				return err
			}
			rep = &ProductResponse{Data: a}
			err = stream.Send(rep)
			if err != nil {
				log.Fatalln(err)
			}
			return nil
		}
		rep := &ProductResponse{ProdStock: int32(count)}

		err := stream.Send(rep)
		if err != nil {
			log.Fatalln(err)
		}
		count--
	}
}

func (p *prodService) SayHelloStream(stream ProdService_SayHelloStreamServer) error {

	wg.Add(1)
	go func() {
		defer wg.Done()

		count := 999
		for {
			time.Sleep(3 * time.Second)
			rsp := &ProductResponse{
				ProdStock: int32(count),
			}
			err := stream.Send(rsp)
			if err != nil {
				return
			}
			count--
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			recv, err := stream.Recv()
			if err != nil {
				return
			}
			fmt.Println("服务端收到客户端消息；", recv.ProdId)
		}
	}()
	wg.Wait()
	return nil
}

func (p *prodService) GetStockById(id int32) int32 {
	return id
}
