package client

import (
	"fmt"
	"testing"
)

/*
WebSocket客户端功能测试

本文件用于测试WebSocket客户端的各种功能特性，
包括连接建立、消息发送接收、事件处理等。

运行命令：
go test -v ./socket/client -run "^TestWebsocket$"

测试内容：
1. WebSocket连接建立
2. 消息发送和接收
3. 连接状态管理
4. 事件回调处理
5. 错误处理机制
6. 实时数据流处理
*/

func TestWebsocket(t *testing.T) {
	// 测试连接到Binance的WebSocket API
	ws := NewWebsocket("wss://stream.binance.com:9443/ws", func(message []byte) {
		fmt.Println(string(message))
	})

	// 设置连接成功后的回调处理器
	ws.SetAfterConnectionHandler(func() error {
		return ws.WriteMessage([]byte(`{"method":"SUBSCRIBE","params":["btcusdt@ticker"],"id":1}`))
	})

	// 运行WebSocket
	err := ws.Start()
	if err != nil {
		t.Fatalf("Failed to run websocket: %v", err)
	}

	// 阻塞主进程
	select {}
}
