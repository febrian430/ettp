package ettp

import (
	"fmt"
	"net"
	"sync"
)

type pool struct {
	address string

	freeConns   []net.Conn
	workerConns []net.Conn

	minIdle int
	maxIdle int

	activeConn int
	maxActive  int

	connMut *sync.Mutex

	jobs chan *queuedJob
}

func newPool(config *ClientConfig) (*pool, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	conns := make([]net.Conn, 0)
	for i := 0; i < config.MinIdle; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return nil, ErrConnectionToServer
		}

		conns = append(conns, conn)
	}

	workerConns := make([]net.Conn, 0)
	for i := 0; i < config.QueueWorkers; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return nil, ErrConnectionToServer
		}

		workerConns = append(workerConns, conn)
	}

	pool := &pool{
		address:     address,
		freeConns:   conns,
		workerConns: workerConns,

		minIdle:    config.MinIdle,
		maxIdle:    config.MaxIdle,
		activeConn: 0,
		maxActive:  config.MaxActive,

		jobs:    make(chan *queuedJob),
		connMut: &sync.Mutex{},
	}

	go pool.handleQueuedJobsAsync()

	return pool, nil
}

func (p *pool) Do(req *Request) (*Response, error) {
	conn, err := p.getConn()
	if err != nil {
		return nil, err
	}

	if conn != nil {
		res, err := doJob(conn, req)
		p.putConn(conn)

		return res, err
	} else {
		job := queuedJob{
			req:    req,
			result: make(chan queuedJobResult),
		}
		p.jobs <- &job

		result := <-job.result
		return result.response, result.err
	}
}

func (p *pool) getConn() (conn net.Conn, err error) {

	p.connMut.Lock()
	if len(p.freeConns) > 0 {
		lastIdx := len(p.freeConns) - 1

		conn = p.freeConns[lastIdx]
		p.freeConns = p.freeConns[:lastIdx]

		p.activeConn += 1

		p.connMut.Unlock()
	} else if p.activeConn < p.maxActive {
		conn, err = p.openConn()
		if err == nil {
			p.activeConn += 1
		}
		p.connMut.Unlock()
	} else {
		p.connMut.Unlock()
	}

	return
}

func (p *pool) Log() {
	fmt.Printf("active conn: %d, max active: %d\n", p.activeConn, p.maxActive)
}

func (p *pool) putConn(conn net.Conn) {
	if conn == nil {
		return
	}

	p.connMut.Lock()
	defer p.connMut.Unlock()

	if len(p.freeConns) < p.maxIdle {
		p.freeConns = append(p.freeConns, conn)
	} else {
		conn.Close()
	}
}

func (p *pool) openConn() (net.Conn, error) {
	fmt.Println("opening conn")
	p.Log()

	conn, err := net.Dial("tcp", p.address)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrConnectionToServer.Error(), err)
	}

	return conn, nil
}

func (p *pool) work(id int, conn net.Conn) {
	for {
		doQueuedJob(conn, <-p.jobs)
	}
}

func (p *pool) handleQueuedJobsAsync() {
	for i := 0; i < len(p.workerConns); i++ {
		fmt.Println("starting worker ", i)
		go p.work(i, p.workerConns[i])
	}
}
