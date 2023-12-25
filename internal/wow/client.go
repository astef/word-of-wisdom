package wow

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"
)

type Client interface {
	GetChallenge(ctx context.Context, rq *ChallengeRequest) (*ChallengeResponse, error)
	GetQuote(ctx context.Context, rq *QuoteRequest) (*QuoteResponse, error)
}

type ClientConfig struct {
	Address string
}

type client struct {
	addr string
}

func NewClient(cfg *ClientConfig) Client {
	return &client{addr: cfg.Address}
}

func (c *client) GetChallenge(ctx context.Context, rq *ChallengeRequest) (*ChallengeResponse, error) {
	return request[ChallengeResponse](ctx, c.addr, rq)
}

func (c *client) GetQuote(ctx context.Context, rq *QuoteRequest) (*QuoteResponse, error) {
	return request[QuoteResponse](ctx, c.addr, rq)
}

func request[T any](ctx context.Context, addr string, request any) (*T, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	defer func() {
		conn.Close()
	}()

	if err := gob.NewEncoder(conn).Encode(request); err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	response := new(T)
	if err := gob.NewDecoder(conn).Decode(response); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return response, nil
}
