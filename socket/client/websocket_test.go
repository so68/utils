package client

import (
	"fmt"
	"testing"
)

func TestWebsocket(t *testing.T) {
	// 测试连接到Binance的WebSocket API
	ws := NewWebsocket("wss://stream.binance.com:9443/ws", func(message []byte) {
		fmt.Println(string(message))
	})

	// 设置连接成功后的回调处理器
	ws.SetAfterConnectionHandler(func(websocket *Websocket) error {
		fmt.Println("Connected")
		websocket.WriteMessage([]byte(`{"method":"SUBSCRIBE","params":["btcusdt@ticker"],"id":1}`))
		return nil
	})

	// 运行WebSocket
	err := ws.Start()
	if err != nil {
		t.Fatalf("Failed to run websocket: %v", err)
	}

	// 阻塞主进程
	select {}
}
