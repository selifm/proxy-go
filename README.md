# 内网穿透工具

- 多协程与通道配合达到快速响应
- 3秒发送一次心跳包维护连接
- client断开自动重连
- server断开自动重连

## 说明
- server端： 具有公网地址的服务器
- client端： 需要内网穿透的主机

## 使用

server端, 默认本地为5200端口
```bash
./server -l 5200 -r 3333
```

client端
```bash
./client -l 8080 -r 3333 -h 公网IP地址
```
server 与 client 通过端口3333 实现tcp连接 
client 与 内网中的 `8080`端口程序（需要穿透的服务程序）实现tcp连接

用户访问 `公网IP地址:5200` 即可访问到 内网中的 `8080`端口程序

user_server端, 模拟用户家里公司电脑（需要穿透提供服务的server）
```bash
./user_server
```

user端, 模拟用户请求,控制台模拟输入请求内容
```bash
./user
```

程序定制 qq:455697968
