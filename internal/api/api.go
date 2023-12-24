package api

import "encoding/gob"

type ChallengeRequest struct {
}

type ChallengeResponse struct {
	Challenge []byte
	Signature []byte
}

type QuoteRequest struct {
}

type QuoteResponse struct {
	Quote string
}

func RegisterContracts() {
	// required for encoding/decoding to/from interface work
	gob.Register(&ChallengeRequest{})
	gob.Register(&ChallengeResponse{})
	gob.Register(&QuoteRequest{})
	gob.Register(&QuoteResponse{})
}
