package main

import (
	"net"
	"strings"
)
type User struct{
	Name string
	Addr string
	C chan string
	conn net.Conn
	server *Server
}
//创建一个用户的api
func  NewUser(conn net.Conn,server *Server) *User{
	userAddr:=conn.RemoteAddr().String()
	user:=&User{
		Name: userAddr,
		Addr: userAddr,
		C:make(chan string),
		conn:conn,
		server: server,
	}
	//启动当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}
//用户上线业务
func(this *User)Online(){
	//用户上线，将用户加入表中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name]=this
	this.server.mapLock.Unlock()
	//广播当前用户上线信息
	this.server.BroadCast(this,"it is appeared")

}
//用户下线业务
func(this *User)Offline(){
	//用户下线，将用户从表中消除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap,this.Name)
	this.server.mapLock.Unlock()
	//广播当前用户下线信息
	this.server.BroadCast(this,"it is disappeared")

}

func (this *User)SendMsg(msg string){
	this.conn.Write([]byte(msg))
}
//用户处理消息的业务
func(this *User)Deal(msg string){
	if msg =="who"{
		//查询当前在线用户
		this.server.mapLock.Lock()
		for _,user := range this.server.OnlineMap{
			OnlineMsg:="["+user.Addr+"]"+user.Name+":"+"online..."
			this.SendMsg(OnlineMsg)
		}
		this.server.mapLock.Unlock()
	}else if len(msg)>7&&msg[:7]=="rename|"{
		newName :=strings.Split(msg,"|")[1]
		//判断name是否被占用
		_, ok:=this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("this name already exists"+"\n")
		}else{
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName]=this
			this.server.mapLock.Unlock()
			this.Name=newName
			this.SendMsg("update successfully"+"\n")
		}
	}else if len(msg)>3&&msg[:3]=="to|"{

        //获取用户名
	     remoteName:=strings.Split(msg,"|")[1]
		 if remoteName ==""{
			this.SendMsg("err!!!,please use <to|xxx|i fuck you>\n")
			return
		 }
		//根据用户名得到server对象
		 remoteUser, ok := this.server.OnlineMap[remoteName]
		 if !ok{
			this.SendMsg("not have this user\n")
			return
		 }
		 //获取消息内容，通过对方的user对象将消息内容发送过去
		 content:=strings.Split(msg,"|")[2]
		 if content==""{
			this.SendMsg("don't have content,please sent again\n")
			return
		 }
		 remoteUser.SendMsg(this.Name+"say to you:"+content)


	}else{

	//将得到的消息进行广播
	this.server.BroadCast(this,msg)
	}
	

}
func(this *User)ListenMessage(){

	for{
		msg:=<-this.C
		this.conn.Write([]byte(msg+"\n"))
	}
}