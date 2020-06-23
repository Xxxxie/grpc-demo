package main

import(
    "log"
    "os"
    "context"
    "google.golang.org/grpc"
    "grpc-demo/api"
)

const (
    address ="localhost:22221"
    defaultName = "world"
)

func main() {
    conn, err := grpc.Dial(address, grpc.WithInsecure())

    if err != nil{
        log.Fatalf("did not connect: %v", err)
    }

    defer conn.Close()

    c := api.NewGreeterClient(conn)

    name := defaultName

    if len(os.Args) >1{
        name = os.Args[1]
    }

    req := api.HelloRequest{
        Name:name
    }

    reps, err := c.SayHello(context.Background(), &req)

    if err != nil{
        log.Fatalf("could not greet: %v", err)
    }

    log.Printf("get server Greeting response: %s", reps.Message)
}