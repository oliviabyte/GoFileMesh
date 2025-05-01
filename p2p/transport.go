package p2p

import (
    "net" // 导入网络操作相关的包
    "fmt" // 导入格式化输出相关的包
)

// TCPTransport 结构体定义了TCP传输层的基本属性
type TCPTransport struct {
    listenAddr string // 监听地址（IP:端口）
    listener   net.Listener // TCP监听器
}

// 创建新的TCPTransport实例的工厂函数
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
    t.listener = ln // 保存监听器引用
    fmt.Println("Listening on", t.listenAddr) // 打印监听地址
    go t.acceptLoop() // 启动一个新的goroutine处理连接请求
    return nil // 成功返回nil
}

// acceptLoop方法持续接受新的连接请求
func (t *TCPTransport) acceptLoop() {
    for { // 无限循环，持续接受连接
        conn, err := t.listener.Accept() // 接受新连接
        if err != nil {
            fmt.Println("Accept error:", err) // 打印错误信息
            continue // 继续下一次循环
        }
        fmt.Println("New connection from", conn.RemoteAddr()) // 打印连接来源
        // TODO: handle message
        conn.Close()
    }
}
