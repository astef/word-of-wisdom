package main

import (
	"context"
	"fmt"
	"log"

	"github.com/astef/word-of-wisdom/internal/api"
)

type message struct {
}

type handler struct {
	logDebug *log.Logger
	logInfo  *log.Logger
	logWarn  *log.Logger
	logErr   *log.Logger
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
