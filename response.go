package ettp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const ()

type Response struct {
	Code    int
	Message string
	Data    protoreflect.ProtoMessage

	body []byte
}

// format is {code}~~~~~{message}~~~~~{payload}
func getResponse(resBytes []byte) (*Response, error) {
	msgs := bytes.Split(resBytes, messageDelim)

	if len(msgs) != 3 {
		return nil, ErrInvalidMessage
	}

	code, err := strconv.Atoi(string(msgs[0]))
	if err != nil {
		return nil, errors.New("response code not int")
	}

	responseMsg := msgs[1]
	if bytes.Equal(responseMsg, nilValue) {
		responseMsg = []byte(" ")
	}

	var body []byte = nil
	if !bytes.Equal(msgs[2], nilValue) {
		body = msgs[2]
	}

	return &Response{
		Code:    code,
		Message: string(responseMsg),
		body:    body,
	}, nil
}

// TODO: refactor
func (res Response) BindProto(dest protoreflect.ProtoMessage) error {
	if res.body == nil || len(res.body) == 0 {
		return nil
	}

	err := proto.Unmarshal(res.body, dest)
	if err != nil {
		return WrapErr("error unmarshal proto response", err)
	}

	return nil
}

func (res Response) Format() (string, error) {
	var err error

	jsonBytes := nilValue
	if res.Data == nil {
		return fmt.Sprintf("%d%s%s%s%v", res.Code, messageDelim, res.Message, messageDelim, string(jsonBytes)), nil
	}

	jsonBytes, err = proto.Marshal(res.Data)
	if err != nil {
		return "", WrapErr("error marshal data to json", err)
	}

	return fmt.Sprintf("%d%s%s%s%v", res.Code, messageDelim, res.Message, messageDelim, string(jsonBytes)), nil
}
