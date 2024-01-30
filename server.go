package ettp

import (
	"fmt"
	"log"
	"net"
)

type HandlerFunc func(req *Request) *Response

type Server struct {
	actions map[string]HandlerFunc
}

func NewServer() *Server {
	return &Server{
		actions: make(map[string]HandlerFunc),
	}
}

func (srv *Server) Handle(action string, handlerFunc HandlerFunc) {
	srv.actions[action] = handlerFunc
}

func (srv Server) ListenAndServe(port int) {
	address := fmt.Sprintf("%s:%d", "127.0.0.1", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("error listening ", err)
	}

	fmt.Println("LISTENING ON", address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("error connection: ", err)
			return
		}

		go srv.handleConn(conn)
	}
}

func (srv Server) handleConn(c net.Conn) {
	// defer c.Close()
	for {
		req, err := getRequest(c)
		if err != nil {
			return
		}

		handlerFunc, ok := srv.actions[req.Action]
		if !ok {
			log.Println(ErrActionNotFound)
			return
		}

		res := handlerFunc(req)

		formattedResponse, _ := res.Format()
		if err != nil {
			res = &Response{
				Code:    999,
				Message: "error formatting response",
				Data:    nil,
			}
			formattedResponse, _ = res.Format()
		}

		log.Printf("Action hit: %s\nResponse: %s\n", req.Action, formattedResponse)

		resBytes := append([]byte(formattedResponse), endValue...)
		_, err = c.Write(resBytes)
		if err != nil {
			log.Printf("error writing response '%s' : %v", resBytes, err)
		}
	}
}
