package main

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"unsafe"
)

var i uint32 = 0

func main() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Panic(err)
	}
	config := tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair("./ssl/server.crt", "./ssl/server.key")
	if err != nil {
		log.Fatalf("load certFile or keyFile failed!err:=%v", err)
		return
	}
	config.PreferServerCipherSuites = true
	config.NextProtos = append(config.NextProtos, "http/1.1")
	tlsListener := tls.NewListener(ln.(*net.TCPListener), &config)

	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			log.Println("Accept err:", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		fmt.Println("this is not tls conn!")
		return
	}
	if err := tlsConn.Handshake(); err != nil {
		fmt.Printf("http: TLS handshake error from %s: %v", tlsConn.RemoteAddr(), err)
		return
	}

	content := make([]byte, 1024)
	_, err := tlsConn.Read(content)
	gi := atomic.AddUint32(&i, 1)
	log.Printf("g%d content:%s", gi, string(content))
	if err != nil {
		log.Printf("g%d error:%s", gi, err.Error())
		return
	}
	if string(content[0:3]) == "GET" {
		headers := parseHandshake(string(content))
		log.Println("headers", headers)
		secWebsocketKey := headers["Sec-WebSocket-Key"]

		// NOTE：这里省略其他的验证
		guid := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

		// 计算Sec-WebSocket-Accept
		h := sha1.New()

		io.WriteString(h, secWebsocketKey+guid)
		accept := make([]byte, 28)
		base64.StdEncoding.Encode(accept, h.Sum(nil))

		response := "HTTP/1.1 101 Switching Protocols\r\n"
		response = response + "Sec-WebSocket-Accept: " + string(accept) + "\r\n"
		response = response + "Connection: Upgrade\r\n"
		response = response + "Upgrade: websocket\r\n\r\n"

		if n, err := tlsConn.Write([]byte(response)); err != nil {
			log.Println("write response failed!err:=%v", err)
		} else {
			fmt.Println("write response(%d) success!", n)
		}
	}
	wssocket := NewWsSocket(tlsConn)
	for {
		_, err := wssocket.ReadIframe()
		if err != nil {
			log.Println("readIframe err:", err)
			return
		}
	}
}

func parseHandshake(content string) map[string]string {
	headers := make(map[string]string, 10)
	lines := strings.Split(content, "\r\n")

	for _, line := range lines {
		if len(line) >= 0 {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				headers[strings.Trim(words[0], " ")] = strings.Trim(words[1], " ")
			}
		}
	}
	return headers
}

type WsSocket struct {
	MaskingKey []byte
	Conn       *tls.Conn
}

func NewWsSocket(conn *tls.Conn) *WsSocket {
	return &WsSocket{Conn: conn}
}

func (this *WsSocket) SendIframe(data []byte, opcode int) error {
	length := len(data)
	maskedData := make([]byte, length)
	for i := 0; i < length; i++ {
		if this.MaskingKey != nil {
			maskedData[i] = data[i] ^ this.MaskingKey[i%4]
		} else {
			maskedData[i] = data[i]
		}
	}

	var extralenbytes []byte
	if length > 125 && length <= 65535 {
		var lc *[4]byte = (*[4]byte)(unsafe.Pointer(&length))
		extralenbytes = append(extralenbytes, lc[1])
		extralenbytes = append(extralenbytes, lc[0])
		length = 126
	} else if length > 65535 {
		var lc *[4]byte = (*[4]byte)(unsafe.Pointer(&length))
		extralenbytes = append(extralenbytes, byte(0))
		extralenbytes = append(extralenbytes, byte(0))
		extralenbytes = append(extralenbytes, byte(0))
		extralenbytes = append(extralenbytes, byte(0))
		extralenbytes = append(extralenbytes, lc[3])
		extralenbytes = append(extralenbytes, lc[2])
		extralenbytes = append(extralenbytes, lc[1])
		extralenbytes = append(extralenbytes, lc[0])
		length = 127
	}
	fmt.Println("send extrabytes:", extralenbytes, "opcode:", opcode)

	_, err := this.Conn.Write([]byte{byte(128 | opcode)})
	if err != nil {
		return err
	}
	if length <= 0 {
		return nil
	}

	var payLenByte byte
	if this.MaskingKey != nil && len(this.MaskingKey) != 4 {
		payLenByte = byte(0x80) | byte(length)
		_, err = this.Conn.Write([]byte{payLenByte})
		if err != nil {
			return err
		}
		if len(extralenbytes) > 0 {
			_, err = this.Conn.Write(extralenbytes)
			if err != nil {
				return err
			}
		}
		_, err = this.Conn.Write(this.MaskingKey)
		if err != nil {
			return err
		}
	} else {
		payLenByte = byte(0x00) | byte(length)
		_, err = this.Conn.Write([]byte{payLenByte})
		if err != nil {
			return err
		}
		if len(extralenbytes) > 0 {
			_, err = this.Conn.Write(extralenbytes)
			if err != nil {
				return err
			}
		}
	}
	_, err = this.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (this *WsSocket) ReadIframe() (data []byte, err error) {
	err = nil
	opcodeByte := make([]byte, 1)
	_, err = this.Conn.Read(opcodeByte)
	if err != nil {
		return
	}
	FIN := opcodeByte[0] >> 7
	RSV1 := opcodeByte[0] >> 6 & 1
	RSV2 := opcodeByte[0] >> 5 & 1
	RSV3 := opcodeByte[0] >> 4 & 1
	OPCODE := opcodeByte[0] & 15
	log.Println(RSV1, RSV2, RSV3, OPCODE)
	if int(OPCODE) == 8 {
		fmt.Println("close the connect!")
		err = fmt.Errorf("connection closed")
		return
	} else if int(OPCODE) == 9 {
		fmt.Println("recv ping frame!")
		err = this.SendIframe(nil, 10)
		return
	}
	payloadLenByte := make([]byte, 1)
	this.Conn.Read(payloadLenByte)
	if err != nil {
		return
	}
	payloadLen := uint64(payloadLenByte[0] & 0x7F)
	mask := payloadLenByte[0] >> 7
	if payloadLen == 126 {
		extendedByte := make([]byte, 2)
		this.Conn.Read(extendedByte)
		if err != nil {
			return
		}
		var lc *[8]byte = (*[8]byte)(unsafe.Pointer(&payloadLen))
		(*lc)[1] = extendedByte[0]
		(*lc)[0] = extendedByte[1]
	} else if payloadLen == 127 {
		extendedByte := make([]byte, 8)
		this.Conn.Read(extendedByte)
		if err != nil {
			return
		}
		var lc *[8]byte = (*[8]byte)(unsafe.Pointer(&payloadLen))
		(*lc)[7] = extendedByte[0]
		(*lc)[6] = extendedByte[1]
		(*lc)[5] = extendedByte[2]
		(*lc)[4] = extendedByte[3]
		(*lc)[3] = extendedByte[4]
		(*lc)[2] = extendedByte[5]
		(*lc)[1] = extendedByte[6]
		(*lc)[0] = extendedByte[7]
	}

	maskingByte := make([]byte, 4)
	if mask == 1 {
		this.Conn.Read(maskingByte)
		if err != nil {
			return
		}
		this.MaskingKey = maskingByte
	}
	payloadDataByte := make([]byte, payloadLen)
	this.Conn.Read(payloadDataByte)
	if err != nil {
		return
	}
	dataByte := make([]byte, payloadLen)
	for i := uint64(0); i < payloadLen; i++ {
		if mask == 1 {
			dataByte[i] = payloadDataByte[i] ^ maskingByte[i%4]
		} else {
			dataByte[i] = payloadDataByte[i]
		}
	}
	if FIN == 1 {
		data = dataByte
		err = this.SendIframe(data, 1)
		return
	}

	nextData, err := this.ReadIframe()
	if err != nil {
		return
	}
	data = append(data, nextData...)
	return
}
