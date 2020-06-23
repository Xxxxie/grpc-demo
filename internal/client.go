package main

import(
    "log"
    "os"
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "grpc-demo/api"
)

const (
    address ="localhost:22221"
    defaultName = "world"

    // TLS认证选项
    OpenTLS = true
)

// customCredential 自定义认证
type customCredential struct {}

func (c customCredential) GetRequestMetadata(ctx context.Context, uri ... string) (map[string]string, error){
    return map[string]string {
        "appid":"110",
        "appkey": "key",
    },nil
}

func (c customCredential) RequireTransportSecurity() bool{
    if OpenTLS{
        return true
    }

    return false;
}

func main() {

    var opts []grpc.DialOption

    if OpenTLS{
        // 添加证书
        creds, err := credentials.NewClientTLSFromFile("../keys/server.pem","hello")
        if err !=nil {
            log.Fatalf("Failed to create TLS credentials %v", err)
        }
        opts = append(opts, grpc.WithTransportCredentials(creds))
    }else{
        opts = append(opts, grpc.WithInsecure())
    }

    // 添加自定义token认证
    opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

    // 构建connect
    conn, err := grpc.Dial(address, opts...)
    if err != nil{
        log.Fatalf("did not connect: %v", err)
    }

    // 延时函数 在return之前进行
    defer conn.Close()

    // 初始化客户端
    c := api.NewGreeterClient(conn)

    // 获取客户名称
    name := defaultName
    if len(os.Args) >1{
        name = os.Args[1]
    }

    // 构造请求
    req := api.HelloRequest{
        Name:name,
    }

    // 远程SayHello函数
    reps, err := c.SayHello(context.Background(), &req)
    if err != nil{
        log.Fatalf("could not greet: %v", err)
    }

    // 结果
    log.Printf("get server Greeting response: %s", reps.Message)
}