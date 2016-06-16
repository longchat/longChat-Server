package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"unsafe"
)

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
	Conn       net.Conn
}

func NewWsSocket(conn net.Conn) *WsSocket {
	return &WsSocket{Conn: conn}
}

func (this *WsSocket) HandShake() error {
	content := make([]byte, 1024)
	_, err := this.Conn.Read(content)
	if err != nil {
		log.Printf("read content error:%s", err.Error())
		return err
	}
	if string(content[0:3]) == "GET" {
		headers := parseHandshake(string(content))
		secWebsocketKey := headers["Sec-WebSocket-Key"]
		if secWebsocketKey == "" {
			return errors.New("invalid Sec-WebSocket-Key")
		}
		guid := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		h := sha1.New()

		io.WriteString(h, secWebsocketKey+guid)
		accept := make([]byte, 28)
		base64.StdEncoding.Encode(accept, h.Sum(nil))

		response := "HTTP/1.1 101 Switching Protocols\r\n"
		response = response + "Sec-WebSocket-Accept: " + string(accept) + "\r\n"
		response = response + "Connection: Upgrade\r\n"
		response = response + "Upgrade: websocket\r\n\r\n"

		if _, err := this.Conn.Write([]byte(response)); err != nil {
			log.Println("write response failed!err:=%v", err)
			return err
		}
	} else {
		return errors.New("unsupported request!")
	}
	return nil
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
	log.Println(FIN, RSV1, RSV2, RSV3, OPCODE)
	if int(OPCODE) == 8 {
		err = errors.New("connection closed")
		return
	} else if int(OPCODE) == 9 {
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
		_, err = this.Conn.Read(extendedByte)
		if err != nil {
			return
		}
		var lc *[8]byte = (*[8]byte)(unsafe.Pointer(&payloadLen))
		(*lc)[1] = extendedByte[0]
		(*lc)[0] = extendedByte[1]
	} else if payloadLen == 127 {
		extendedByte := make([]byte, 8)
		_, err = this.Conn.Read(extendedByte)
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
		_, err = this.Conn.Read(maskingByte)
		if err != nil {
			return
		}
		this.MaskingKey = maskingByte
	}
	payloadDataByte := make([]byte, payloadLen)
	_, err = this.Conn.Read(payloadDataByte)
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
		return
	}

	nextData, err := this.ReadIframe()
	if err != nil {
		return
	}
	data = append(data, nextData...)
	return
}
