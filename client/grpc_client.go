package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gRPCProtobufStudy/client/auth"
	"gRPCProtobufStudy/client/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

func main() {
	//单向认证
	//creds, err2 := credentials.NewClientTLSFromFile("cert/server.pem", "*.mszlu.com")
	//if err2 != nil {
	//	log.Fatal("证书错误", err2)
	//}

	// 证书认证-双向认证
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, _ := tls.LoadX509KeyPair("cert/client.pem", "cert/client.key")
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()
	ca, _ := ioutil.ReadFile("cert/ca.crt")
	// 尝试解析所传入的 PEM 编码的证书。如果解析成功会将其加到 CertPool 中，便于后面的使用
	certPool.AppendCertsFromPEM(ca)
	// 构建基于 TLS 的 TransportCredentials 选项
	creds := credentials.NewTLS(&tls.Config{
		// 设置证书链，允许包含一个或多个
		Certificates: []tls.Certificate{cert},
		// 要求必须校验客户端的证书。可以根据实际情况选用以下参数
		ServerName: "*.mszlu.com",
		RootCAs:    certPool,
	})

	//Token
	user := &auth.Authentication{
		User:     "admin",
		Password: "admin",
	}

	//连接8002
	conn, err := grpc.Dial(":8002", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(user))
	if err != nil {
		log.Fatalln("连接出错", err)
	}
	// 退出时关闭链接
	defer conn.Close()

	// 2. 调用Product.pb.go中的NewProdServiceClient方法
	productServiceClient := service.NewProdServiceClient(conn)

	// 3. 直接像调用本地方法一样调用GetProductStock方法
	request := &service.ProductRequest{
		ProdId: 123,
	}
	resp, err := productServiceClient.GetProductStock(context.Background(), request)
	if err != nil {
		log.Fatal("调用gRPC方法错误: ", err)
	}

	fmt.Println("调用gRPC方法成功，ProdStock = ", resp)

	//客户端流
	stream, err := productServiceClient.UpdateStockClientStream(context.Background())
	if err != nil {
		log.Fatalln("获取流失败")
	}
	rsp := make(chan struct{}, 1)
	go prodRequest(stream, rsp)
	select {
	case <-rsp:
		recv, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalln(err)
		}
		stock := recv.ProdStock
		fmt.Printf("客户端收到响应：", stock)
	}

	//服务端流
	serverStream, err := productServiceClient.GetProductStockServerStream(context.Background(), request)
	if err != nil {
		return
	}
	for {
		recv, err := serverStream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println("服务端数据发送完毕")
				err := serverStream.CloseSend()
				if err != nil {
					log.Fatalln(err)
				}
				break
			}
		}
		fmt.Println("服务端发送过来的流：", recv)
	}

	//双向流,心跳检测
	var wg sync.WaitGroup
	doublestream, err := productServiceClient.SayHelloStream(context.Background())
	if err != nil {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for {
			request := &service.ProductRequest{
				ProdId: int32(count),
			}
			err := doublestream.Send(request)
			if err != nil {
				return
			}
			count++
			time.Sleep(time.Second)
		}

	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			recv, err := doublestream.Recv()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("客户端收到的信息：", recv.ProdStock)

		}
	}()
	wg.Wait()
}

func prodRequest(stream service.ProdService_UpdateStockClientStreamClient, rsp chan struct{}) {
	count := 0
	for {
		request := &service.ProductRequest{
			ProdId: 123,
		}
		err := stream.Send(request)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second)
		count++
		if count > 10 {
			rsp <- struct{}{}
			return
		}
	}
}
