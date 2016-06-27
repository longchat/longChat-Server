package websocket

import (
	"net"
	"testing"
	"time"
)

func startServer(t *testing.T) {
	ln, err := net.Listen("tcp", ":8899")
	if err != nil {
		t.Fatalf("cant listen at port 8899,err:=%v", err)
	}
	for {
		sconn, err := ln.Accept()
		if err != nil {
			t.Fatal("accept conn failed,err:=%v", err)
		}
		go func() {
			defer sconn.Close()
			wssocket := NewWsSocket(sconn)
			err = wssocket.HandShake()
			if err != nil {
				t.Fatalf("handshake from server failed,err:=%v", err)
			}
			for {
				data, err := wssocket.ReadIframe()
				if err != nil {
					if err.Error() == "connection closed" {
						return
					}
					t.Fatalf("readIframe from server failed,err:=%v", err)
				}
				err = wssocket.SendIframe(data, 1)
				if err != nil {
					t.Fatalf("sendIframe from server failed,err:=%v", err)
				}
			}
		}()
	}
}

func TestWesocketSendAndReadIFrame(t *testing.T) {
	go startServer(t)

	time.Sleep(time.Millisecond * 10)
	cconn, err := net.Dial("tcp", "127.0.0.1:8899")
	if err != nil {
		t.Fatalf("dial 127.0.0.1:8899 failed,err:=%v", err)
	}
	defer cconn.Close()
	_, err = cconn.Write([]byte("GET / HTTP/1.1\r\nHost: 192.168.5.32:8000\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: 7Tuu1RX0QgkhwCV4sEKTCQ==\r\n\r\n"))
	if err != nil {
		t.Fatalf("write header failed,err:=%v", err)
	}
	content := make([]byte, 1024)
	_, err = cconn.Read(content)
	if err != nil {
		t.Fatalf("read reponse failed,err:=%v", err)
	}
	header := parseHandshake(string(content))
	if header["Connection"] != "Upgrade" || header["Upgrade"] != "websocket" {
		t.Fatalf("handshake response invalid!content:%s", string(content))
	}
	cws := NewWsSocket(cconn)
	originData := "hello"
	err = cws.SendIframe([]byte(originData), 1)
	if err != nil {
		t.Fatalf("sendIframe from client failed,err:=%v", err)
	}
	data, err := cws.ReadIframe()
	if err != nil {
		t.Fatalf("readIframe from client failed,err:=%v", err)
	}
	if string(data) != originData {
		t.Fatalf("readIframe from client data not matched!origindata:%s,recvdata:%s", originData, data)
	}
	//close the websocket connection
	err = cws.SendIframe(nil, 8)
	if err != nil {
		t.Fatalf("sendIframe from client failed,err:=%v", err)
	}
}
