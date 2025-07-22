package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string,serverPort int)*Client{
	//创建client对象
	client :=&Client{
		ServerIp: serverIp,
		ServerPort: serverPort,
		flag: 999,

	}
	//连接server
	conn,err:=net.Dial("tcp",fmt.Sprintf("%s:%d",serverIp,serverPort))
	if err !=nil{
		fmt.Println("net.Dial error:",err)
		return nil
	}
	client.conn=conn
	//返回对象
	return client
}


//处理server回应的消息，直接显示到标准输出即可
func(client *Client)dealresponse(){
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout,client.conn)
	
}

//菜单
func (client *Client)menu() bool{
	var flag int
	fmt.Print("1:public\n")
	fmt.Print("2:private\n")
	fmt.Print("3:update name\n")
	fmt.Print("0:exit\n")

	fmt.Scanln(&flag)
	if flag >=0 && flag<=3{
		client.flag=flag
		return  true

	}else{
		fmt.Print("please do right choose")
		return  false
	}
}

//更新用户名
func(client *Client)UpdateName()bool{
	fmt.Print(">>>please new name")
	fmt.Scanln(&client.Name)
	sendmsg:="rename|"+client.Name+"\n"
	_, err:=client.conn.Write([]byte(sendmsg))
	if err!=nil{
		fmt.Print("conn.write err!",err)
		return false
	}
	return true
}
//公共聊天
func (client *Client) PublicChat() bool {
	var chat string
	fmt.Print("请输入公聊内容(直接回车取消): ")
	// 使用ReadLine读取整行，包括空格
	reader := bufio.NewReader(os.Stdin)
	chat, _ = reader.ReadString('\n')
	chat = strings.TrimSpace(chat)
	if chat == "" {
		return true // 空消息视为取消
	}

	sendmsg := "public|" + client.Name + ":" + chat + "\n"
	_, err := client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.write err:", err)
		return false
	}
	return true
}

// 私聊
func (client *Client) PrivateChat() bool {
	// 先获取在线用户列表
	_, err := client.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("获取用户列表失败:", err)
		return false
	}

	var targetName, chat string
	fmt.Print("请输入私聊对象用户名(直接回车取消): ")
	fmt.Scanln(&targetName)
	if targetName == "" {
		return true // 空用户名视为取消
	}

	fmt.Printf("请输入对%s说的内容(直接回车取消): ", targetName)
	// 使用ReadLine读取整行，包括空格
	reader := bufio.NewReader(os.Stdin)
	chat, _ = reader.ReadString('\n')
	chat = strings.TrimSpace(chat)
	if chat == "" {
		return true // 空消息视为取消
	}

	sendmsg := "to|" + targetName + "|" + client.Name + ":" + chat + "\n"
	_, err = client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.write err:", err)
		return false
	}
	return true
}

func(client *Client)Run(){
	for client.flag!=0{
		for client.menu()!=true{

		}
		//根据不同模式处理不同的业务
		switch client.flag{
		case 1:
			//public
		    client.PublicChat()
			break
		case 2:
			//pravite
			client.PrivateChat()
			break
		case 3:
			//update name
			client.UpdateName()
			break
		
		}
	}
}


var serverIp string
var serverPort int
func init(){
	flag.StringVar(&serverIp,"ip","127.0.0.1","设置服务器ip地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort,"port",8888,"设置服务器端口（默认是8888)")
}
func main(){
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp,serverPort)
	if client == nil{
		fmt.Print(">>>>>link to server=lose")
		return 
	}
	//单独开启一个gorotine去处理server的回信
	go client.dealresponse()
	fmt.Println(">>>>>link to server=success")
	//启动客户端业务
	client.Run()
}