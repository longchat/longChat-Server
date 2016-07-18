package message

import (
	"encoding/json"
	"errors"
)

type MessageData struct {
	Data []byte
}

type DataType byte

const (
	DataTypeUserMessage     DataType = 49
	DataTypeGroupMessage    DataType = 50
	DataTypeGroupMemberList DataType = 51
	DataTypeInvalidType     DataType = 126
)

type DataUserMessage struct {
	Id      string
	From    string
	To      string
	Content string
	Type    string
}

type DataGroupMessage struct {
	Id      string
	From    string
	GroupId string
	Content string
	Type    string
}

type DataGroupMembers struct {
	GroupId int64
	UserIds []string
}

type DataGroupMemberUpdate struct {
	UserId  string
	GroupId string
	Type    string
}

func (s *MessageData) GetHeader() DataType {
	if len(s.Data) < 1 {
		return DataTypeInvalidType
	}
	return DataType(s.Data[0])
}

func (s *MessageData) GetData(data interface{}) error {
	if len(s.Data) < 2 {
		return errors.New("data invalid")
	}
	err := json.Unmarshal(s.Data[1:], data)
	if err != nil {
		return err
	}
	return nil
}

func (s *MessageData) Serialize(Type DataType, data interface{}) ([]byte, error) {
	if data == nil {
		return []byte{byte(Type)}, nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, len(b)+1)
	bs[0] = byte(Type)
	copy(bs[1:], b)
	return bs, nil
}
