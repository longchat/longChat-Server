package moniter

//全局的聊天节点树，记录当前整个聊天系统的状态
type MsgNode struct {
	Address string
	State   int
	//Level为服务器默认所在层级，0代表根节点，以此类推
	Level  int
	Child  []*MsgNode
	Parent *MsgNode
}

func StartMoniter() {

}
