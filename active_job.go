package ettp

import (
	"bytes"
	"fmt"
	"net"
)

type queuedJobResult struct {
	response *Response
	err      error
}

type queuedJob struct {
	req    *Request
	result chan queuedJobResult
}

func doJob(conn net.Conn, req *Request) (response *Response, err error) {
	if len(req.Payload) == 0 {
		req.Payload = string(nilValue)
	}

	msg := fmt.Sprintf("%s%s%v", req.Action, messageDelim, req.Payload)

	// fmt.Printf("writing to conn !!%s!!\n", msg)
	// fmt.Printf("writing to conn bytes '%v'\n", []byte(msg))

	//add endValue to end of request
	msgBytes := append([]byte(msg), endValue...)

	_, err = conn.Write(msgBytes)
	if err != nil {
		err = WrapErr("sending request", err)
		return
	}

	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		err = WrapErr("read message from server", err)
		return
	}

	buffer = bytes.Split(buffer, endValue)[0]

	// buffer = bytes.Trim(buffer, "\x00")
	response, err = getResponse(buffer)
	if err != nil {
		err = WrapErr("parsing response", err)
		return
	}

	return
}

func doQueuedJob(conn net.Conn, job *queuedJob) {
	res, err := doJob(conn, job.req)
	job.result <- queuedJobResult{
		response: res,
		err:      err,
	}
}
