package message

var groupUser map[int64]map[int64]struct{}

var userConn map[int64][]*Conn

var command chan interface{}

type connAdd struct {
	conn    *Conn
	userId  int64
	groupId []int64
}

type connDel struct {
	conn    *Conn
	userId  int64
	groupId []int64
}

func controller() {
	connId := 0
	userConn = make(map[int64][]*Conn, 500)
	groupUser = make(map[int64]map[int64]struct{}, 100)
	command = make(chan interface{})
	for {
		select {
		case item := <-command:
			switch value := item.(type) {
			case connAdd:
				addConn(value, &connId)
			case connDel:
				removeConn(value)
			}
		}
	}
}
func removeConn(c connDel) {
}
func addConn(c connAdd, id *int) {
	(*id)++
	c.conn.id = *id
	user, isok := userConn[c.userId]
	if !isok {
		user = make([]*Conn, 1)
		user[0] = c.conn
	} else {
		user = append(user, c.conn)
	}
	userConn[c.userId] = user
	if !isok {
		group, isok := groupUser[c.userId]
		if !isok {
			group = make(map[int64]struct{}, 20)
		}
		var a struct{}
		group[c.userId] = a
		groupUser[c.userId] = group
	}
}
