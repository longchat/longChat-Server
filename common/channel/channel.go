package channel

import (
	"net"

	"github.com/longchat/longChat-Server/common/channel/protoc"
)

var chans map[[4]byte]chan protoc.Message

func convertIpTo4Byte(ip net.IP) [4]byte {
	var b [4]byte
	b[0] = ip[0]
	b[1] = ip[1]
	b[2] = ip[2]
	b[3] = ip[3]
	return b
}

func GetMsgChannel(ip net.IP) chan protoc.Message {
	return chans[convertIpTo4Byte(ip)]
}
