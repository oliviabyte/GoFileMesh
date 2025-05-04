package p2p

import (
	"fmt" // 导入格式化输出相关的包
	"net" // 导入网络操作相关的包
)

// 进行网络通信的组件，用来建立、监听、接受 TCP 连接（比如和别的节点通信）
// TCPTransport 结构体定义了TCP传输层的基本属性
type TCPTransport struct {
	listenAddr string                             // 监听地址（IP:端口）
	listener   net.Listener                       // TCP监听器
	handlers   map[string]func(Message, net.Conn) // 新增，键类型为字符串，值类型是一个函数
}

// 相当于java构造器
func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

// Start方法启动TCP监听服务
func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr) // 在指定地址上开始监听TCP连接，这是Go中创建网络监听器的标准方式
	if err != nil {
		return err // 如果监听失败，返回错误
	}
	t.listener = ln                           // 保存监听器引用
	fmt.Println("Listening on", t.listenAddr) // 打印监听地址
	go t.acceptLoop()                         // 启动一个新的goroutine处理连接请求
	return nil                                // 成功返回nil
}

// acceptLoop方法持续接受新的连接请求
func (t *TCPTransport) acceptLoop() {
	for { // 无限循环，持续接受连接
		conn, err := t.listener.Accept() // 接受新连接
		// Accept()是一个阻塞式方法，程序会在这里停住，直到有人连进来
		// 一旦有连接，就返回一个 conn，你就可以和这个客户端进行双向通信了
		if err != nil {
			fmt.Println("Accept error:", err) // 打印错误信息
			continue                          // 继续下一次循环
		}
		go func(c net.Conn) {
			defer c.Close()

			msg, err := ReceiveMessage(c)
			if err != nil {
				fmt.Println("Error decoding message:", err)
				return
			}

			handler, ok := t.handlers[msg.Type]
			if !ok {
				fmt.Printf("❌ No handler for message type: %s\n", msg.Type)
				return
			}

			handler(msg, c) // 👈 分发给注册的处理函数
		}(conn)
	}
}

// 用来注册处理函数
// 建立一个映射关系，将不同类型的消息（键/key）与相应的处理函数（值/value）关联起来
// 优点：解耦了网络层（transport）和业务层（store、crypto）
func (t *TCPTransport) OnMessage(msgType string, handler func(Message, net.Conn)) {
	if t.handlers == nil {
		t.handlers = make(map[string]func(Message, net.Conn))
	}
	t.handlers[msgType] = handler
}
