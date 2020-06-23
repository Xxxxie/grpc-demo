package main

import(
    "fmt"
    "context"
    "log"
    "net"
    "grpc-demo/api"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const (
    port = ":22221"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloReply, error){

    fmt.Print("get client request name :" + req.Name)

    return &api.HelloReply{Message : "Hello, " + req.Name},nil
}

func main() {
    lis, err := net.Listen("tcp", port)

    if err !=nil{
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()


    // server结构体中 需包含调用的api 注册服务
    api.RegisterGreeterServer(s,&server{})

    reflection.Register(s)

    if err := s.Serve(lis); err != nil{
        log.Fatalf("failed to serve:%v", err)
    }

}