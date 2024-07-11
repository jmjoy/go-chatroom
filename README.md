# go-chatroom

GO写的Websocket聊天室

## shortcut

[![](https://raw.githubusercontent.com/jmjoy/go-chatroom/master/shortcut.jpg)](https://raw.githubusercontent.com/jmjoy/go-chatroom/master/shortcut.jpg)

## 使用说明

利用go的embed特性，将静态文件也编译进二进制文件，因此只需要一个`go-chatroom`二进制文件即可运行聊天室服务。

加上`-h`参数可以打印帮助信息。

聊天室服务需要两个端口，一个是http端口，默认是10000，一个是websocket端口，默认是10001，可以分别通过`-p`和`-wp`参数进行修改。

如果遇到监听的websocket端口跟浏览器使用的websocket端口不一致的情况（比如做了端口映射），则需要另外指定浏览器使用的websocket端口，
通过`-cwp`参数指定。

因为支持发送图片，因此需要图片上传文件到文件夹，`go-chatroom`会自动创建上传文件夹，默认路径是`./go-chatroom-upload`，可以通过`-ud`参数
按需修改。