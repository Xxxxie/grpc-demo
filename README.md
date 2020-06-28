grpc-demo 是grpc的一款demo程序

```
// 启动服务端
go run interval/server.go

// 启动客户端 opt 为可选项，默认为world
go run interval/client.go <opt>
```

.proto 主要保存grpc在网络传输中的数据格式 

当前实现功能|
--|
简单的grpc通信|
在控制台和log文件同时记录log|
grpc中TSL认证|
Token验证(基于拦截器)|
grpc状态码的使用|
拦截器链|
https转换|