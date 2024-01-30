package ettp

import (
	"bytes"
	"log"
	"net"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	messageDelim = []byte("~~~~~")
	nilValue     = []byte("!~!~!")
	endValue     = []byte("`~`~`~`~`")
)

type Header map[string]string

type Request struct {
	Action  string
	Payload string
}

func (r Request) BindProto(dest protoreflect.ProtoMessage) error {
	if len(r.Payload) == 0 {
		return nil
	}

	err := proto.Unmarshal([]byte(r.Payload), dest)
	if err != nil {
		return WrapErr("error unmarshal proto request", err)
	}

	return nil
}

// format is {action~~~~~{payload}
func getRequest(conn net.Conn) (*Request, error) {
	buffer := make([]byte, 1024)

	_, err := conn.Read(buffer)
	if err != nil {
		return nil, WrapErr("read connection message", err)
	}

	buffer = bytes.Split(buffer, endValue)[0]
	// buffer = bytes.Trim(buffer, "\x00")

	// fmt.Printf("GOT REQUEST BYTES '%v'\n", buffer)
	// fmt.Printf("GOT REQUEST !!%s!!\n", string(buffer))

	msgs := bytes.Split(buffer, messageDelim)

	// fmt.Println("BYTES PAYLOAD", msgs[1])
	// fmt.Println("STRING PAYLOAD", string(msgs[1]))
	// fmt.Println("BYTES STRINGED PAYLOAD", []byte(string(msgs[1])))

	// msgs := strings.Split(msg, messageDelim)

	if len(msgs) != 2 {
		return nil, ErrInvalidMessage
	}

	var payload []byte = nil
	if !bytes.Equal(msgs[1], nilValue) {
		payload = msgs[1]
	}

	log.Printf("handling request: '%s'", string(buffer))

	return &Request{
		Action:  string(msgs[0]),
		Payload: string(payload),
	}, nil
}
