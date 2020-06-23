package service

import(
    "log"
    "context"
    "grpc-demo/api"
)


type Helloserver struct{}

func NewHelloService() *Helloserver {
	service := &Helloserver{}
	return service
}

func (s *Helloserver) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloReply, error){

    log.Println("get client request name :" + req.Name)
    resp := new(api.HelloReply)
    resp.Message = "Hello, " + req.Name;

    return resp,nil
}