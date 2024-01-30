package ettp

type Client struct {
	pool *pool
}

type ClientConfig struct {
	Host string
	Port int

	MinIdle   int
	MaxIdle   int
	MaxActive int

	QueueWorkers int
}

func NewClient(config *ClientConfig) (*Client, error) {

	if config.MinIdle <= 1 {
		config.MinIdle = 1
	}

	if config.MaxIdle <= 1 {
		config.MaxIdle = 1
	}

	if config.MaxActive <= 1 {
		config.MaxActive = 1
	}

	if config.QueueWorkers <= 1 {
		config.QueueWorkers = 1
	}

	pool, err := newPool(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		pool: pool,
	}, nil
}

func (c Client) Do(req Request) (*Response, error) {
	return c.pool.Do(&req)
}
