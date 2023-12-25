package main

import (
	"context"
	"fmt"

	"github.com/astef/word-of-wisdom/internal/api"
	"github.com/astef/word-of-wisdom/internal/log"
)

type handler struct {
	logger log.Logger
}

func (h *handler) handle(ctx context.Context, rq any) (any, error) {
	switch v := rq.(type) {
	case *api.ChallengeRequest:
		return h.requestChallenge(ctx, v)
	case *api.QuoteRequest:
		return h.requestQuote(ctx, v)
	default:
		return nil, fmt.Errorf("unknown request type: %T", v)
	}
}

func (h *handler) requestChallenge(ctx context.Context, rq *api.ChallengeRequest) (*api.ChallengeResponse, error) {
	return nil, nil
}

func (h *handler) requestQuote(ctx context.Context, rq *api.QuoteRequest) (*api.QuoteResponse, error) {
	return nil, nil
}
