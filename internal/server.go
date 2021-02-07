package main

import(
    "context"
    "log"
    "net"
    "golang.org/x/net/http2"
    "net/http"
    "crypto/tls"
    "os"
    "io"
    "time"
    "io/ioutil"
    "strings"
    "fmt"

    "grpc-demo/api"
    "grpc-demo/internal/service"
    
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/codes"
)


var logger *log.Logger


func main() {

    var address = "127.0.0.1:22222"
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
    //opts = append(opts, grpc.UnaryInterceptor(InterceptChain(interceptor1, interceptor2)))

    // 在实例化之前添加拦截器
    // 实例化servers
    s := grpc.NewServer(opts...)

    // server结构体中 需包含调用的api 注册服务
    api.RegisterGreeterServer(s, service.NewHelloService())
    
    // http-grpc gateway
    ctx, cancel := context.WithCancel(context.Background())
    
    // WithCancel 返回了临时的cancel函数，用于取消当前的ctx
    defer cancel()

    dcreds, err := credentials.NewClientTLSFromFile("../keys/server.pem", "test")
    if err != nil {
        logger.Fatalf("Failed to create TLS credentials %v", err)
    }

    dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
    gwmux := runtime.NewServeMux()
    err = api.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, address, dopts)
    if err != nil {
        logger.Printf("Failed to register gateway server: %v\n", err)
    }

    // http 服务
    mux := http.NewServeMux()
    mux.Handle("/",gwmux)

    // 开启HTTP服务
    cert, _ := ioutil.ReadFile("../keys/server.pem")
    key, _ := ioutil.ReadFile("../keys/server.key")
    var demoKeyPair *tls.Certificate
    pair,err := tls.X509KeyPair(cert, key)
    if err != nil {
        panic(err)
    }

    demoKeyPair = &pair

    srv := &http.Server{
        Addr : address,
        Handler : grpcHandleFunc(s, mux),
        TLSConfig: &tls.Config{
            Certificates: []tls.Certificate{*demoKeyPair},
            NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
        },
    }

    logger.Printf("grpc and https on port: %d\n", 22222)
    err = srv.Serve(tls.NewListener(lis, srv.TLSConfig))

    reflection.Register(s)

    logger.Println("Listen on " + address + " with TLS + Token + Interceptor")

    if err := s.Serve(lis); err != nil{
        logger.Fatalf("failed to serve:%v", err)
    }

    // 保证log文件关闭
    logFile, err := os.OpenFile(logFilePath,os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    defer logFile.Close()

}

// 初始化log,将log同时输出在log文件和控制台上
func initlog() string {
    file := "../log/" + time.Now().Format("2006-01-02") + ".log"
    fmt.Println(file)
    _,err :=os.Stat(file)
    var f *os.File
    if  err!=nil{
        f, _=os.Create(file)
        fmt.Println(file)
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

     return nil

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
func interceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error){
    
    fmt.Println("Using the interceptor1")
    // 添加Token校验
    //err = auth(ctx)
    //if err != nil {
    //    return
    //}
    // 校验通过后继续处理请求
    return handler(ctx, req)
}

func interceptor2(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error){

    fmt.Println("Using the interceptor2")

    return handler(ctx, req)
}
// 一元拦截器链
func InterceptChain(intercepts... grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor{
    // 获取拦截器长度
    l := len(intercepts)

    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error){

        // 构造一个链
        chain := func(currentInter grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler{
            return func(currentCtx context.Context, currentReq interface{})(interface{}, error){
                return currentInter(currentCtx, currentReq, info, currentHandler)
            }
        }

        // 此处为什么为倒序还未搞明白
        chainHandler := handler
        for i := l-1; i>=0; i-- {
            chainHandler = chain(intercepts[i],chainHandler)
        }

        return chainHandler(ctx,req)
    }
}


// 识别是否为http调用
func grpcHandleFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler{
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){

        if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"),"application/grpc"){
            grpcServer.ServeHTTP(w,r)
        }else{
            otherHandler.ServeHTTP(w,r)
        }
    })
}