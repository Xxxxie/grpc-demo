package main

import(
    "context"
    "log"
    "net"
    "os"
    "io"
    "time"

    "grpc-demo/api"
    "grpc-demo/internal/service"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/codes"
)

const (
    address = "127.0.0.1:22221"
)

var logger *log.Logger

func main() {
    lis, err := net.Listen("tcp", address)
    logFilePath := initlog()
    if err !=nil{
        logger.Fatalf("failed to listen: %v", err)
    }

    // TLS认证
    creds, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")

    if err != nil {
        logger.Fatalf("Failed to generate credentials %v", err)
    }
    

    // 服务选项
    var opts []grpc.ServerOption
    // 添加TLS认证
    opts = append(opts, grpc.Creds(creds))
    // 添加拦截器
    opts = append(opts, grpc.UnaryInterceptor(interceptor))

    // 在实例化之前添加拦截器
    // 实例化servers
    s := grpc.NewServer(opts...)

    // server结构体中 需包含调用的api 注册服务
    api.RegisterGreeterServer(s, service.NewHelloService())

    reflection.Register(s)

    logger.Println("Listen on " + address + " with TLS + Token + Interceptor")

    if err := s.Serve(lis); err != nil{
        logger.Fatalf("failed to serve:%v", err)
    }

    // 保证文件关闭
    logFile, err := os.OpenFile(logFilePath,os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    defer logFile.Close()

}

// 初始化log,将log同时输出在log文件和控制台上
func initlog() string {
    file := "../log/" + time.Now().Format("2006-01-02") + ".log"
    _,err :=os.Stat(file)
    var f *os.File
    if  err!=nil{
        f, _=os.Create(file)
    }else{
        //如果存在文件则 追加log
        f ,_= os.OpenFile(file,os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    }

    writers := []io.Writer{
		f,
        os.Stdout,
    }

    logger = log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime|log.Lshortfile)
    return file
}

// auth 验证token
func auth(ctx context.Context) error{

    // 从ctx中提取metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
    }

    var(
        appid string
        appkey string
    )

    if val, ok := md["appid"]; ok {
        appid = val[0]
    }

    if val, ok := md["appkey"]; ok {
        appkey = val[0]
     }

     if appid != "110" || appkey != "key" {
         return grpc.Errorf(codes.Unauthenticated, "Token认证消息无效：appid=%s, appkey=%s", appid, appkey)
     }

    return nil
}

// 一元服务拦截器 若需要多个拦截器，则需要配置拦截器链，核心思想是递归
/**********************************
/* ctx 上下文
/* req 用户请求的参数
/* info RPC方法的所有信息
/* handler RPC方法本身
/* resp RPC方法执行结果
**********************************/
func interceptor(ctx context.Context, req interface{}, info * grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error){
    
    // 添加Token校验
    err = auth(ctx)
    if err != nil {
        return 
    }
    // 校验通过后继续处理请求
    return handler(ctx, req)
}