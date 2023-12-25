package main

import (
	"context"
	"fmt"

	"github.com/astef/word-of-wisdom/internal/log"
	"github.com/astef/word-of-wisdom/internal/wow"
)

type handler struct {
	logger log.Logger
}

func (h *handler) handle(ctx context.Context, rq any) (any, error) {
	switch v := rq.(type) {
	case *wow.ChallengeRequest:
		return h.requestChallenge(ctx, v)
	case *wow.QuoteRequest:
		return h.requestQuote(ctx, v)
	default:
		return nil, fmt.Errorf("unknown request type: %T", v)
	}
}

func (h *handler) requestChallenge(ctx context.Context, rq *wow.ChallengeRequest) (*wow.ChallengeResponse, error) {
	return nil, nil
}

func (h *handler) requestQuote(ctx context.Context, rq *wow.QuoteRequest) (*wow.QuoteResponse, error) {
	return nil, nil
}
