package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex
	//消息广播的管道
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}
//监听message广播消息channel的goroutine,一旦有消息就发送给全部在线的user
func(this *Server)ListenMessage(){
	for{
		msg :=<-this.Message
		this.mapLock.Lock()
		for _,cli :=range this.OnlineMap{
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func(this*Server)BroadCast(user *User,msg string){
	SendMsg:="<<"+user.Addr+">>"+user.Name+":"+msg+"\n"
	this.Message <- SendMsg
}

func(this *Server)Handler(conn net.Conn){
	//当前业务
	//fmt.Print("建立连接成功")
	user := NewUser(conn,this)
	user.Online()
	
	//监听是否活跃
	isLive:= make(chan bool)
	// 接受用户端发送的消息
    go func() {
            
		buf := make([]byte, 4096)
     
		var inputBuf []byte // 缓存用户输入
     
		for {
			n, err := conn.Read(buf)
			if n == 0 {
                user.Offline()
                return
		    }
            if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
            	return
            }

            // 遍历读取到的字节，寻找回车符
            for i := 0; i < n; i++ {
                if buf[i] == '\r' || buf[i] == '\n' {
                  // 遇到回车，处理完整消息
                  if len(inputBuf) > 0 {
					//提取消息
                     msg := string(inputBuf)
					 //消息处理
                     user.Deal(msg)
					 isLive <- true
                     inputBuf = inputBuf[:0] // 清空缓存
                   }
                   break
                } else {
                // 未到回车，继续缓存
                inputBuf = append(inputBuf, buf[i])}
			}
		}
    }()

	//当前handle阻塞
	for{
		select{
		case <-isLive:
		case <-time.After(time.Second*100):
		       user.SendMsg("you are out")
			   close(user.C)
			   conn.Close()
			   return
		
		}
	}


}
// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	Listener, err := net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip,this.Port))
	if err !=nil{
		fmt.Print("net.listen err:",err)
		return
	}
	//close listen socket
	defer Listener.Close()

	//启动监听message的goroutine
	go this.ListenMessage()
	for{
		//accept
		conn,err :=Listener.Accept()
		if err !=nil{
			fmt.Print("listener accept err",err)
			continue
		}
		//do handler
		go this.Handler(conn)

	}

}