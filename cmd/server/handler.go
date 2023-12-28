package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/astef/word-of-wisdom/internal/log"
	"github.com/astef/word-of-wisdom/internal/wow"
)

type handler struct {
	logger                  log.Logger
	now                     time.Time
	serverSecret            []byte
	clientIP                string
	challengeExpirationSec  int
	challengeDataSize       int
	challengeDifficulty     int
	challengeAvgSolutionNum int
	challengeBlockSize      *big.Int
	cryptoRand              io.Reader
	quoteRandIntn           func(n int) int
}

func (h *handler) handle(rq any) (any, error) {
	switch v := rq.(type) {
	case *wow.ChallengeRequest:
		return h.requestChallenge()
	case *wow.QuoteRequest:
		return h.requestQuote(v)
	default:
		return nil, fmt.Errorf("unknown request type: %T", v)
	}
}

func (h *handler) requestChallenge() (*wow.ChallengeResponse, error) {
	challenge, err := h.generateChallenge()
	if err != nil {
		return nil, err
	}

	signature, err := h.signChallenge(challenge)
	if err != nil {
		return nil, err
	}

	return &wow.ChallengeResponse{
		Challenge: challenge,
		Signature: signature,
	}, nil
}

func (h *handler) requestQuote(rq *wow.QuoteRequest) (*wow.QuoteResponse, error) {
	if err := h.validateSignature(rq.Challenge, rq.Signature); err != nil {
		return nil, err
	}
	if err := h.validateChallenge(rq.Challenge); err != nil {
		return nil, err
	}

	if err := h.validateSolution(rq.Challenge, rq.Solution); err != nil {
		return nil, err
	}

	quoteIndex := h.quoteRandIntn(len(quotes))
	h.logger.Info().Printf("solution is valid, rewarding with quote #%d", quoteIndex)

	return &wow.QuoteResponse{Quote: quotes[quoteIndex]}, nil
}

func (h *handler) generateChallenge() (*wow.Challenge, error) {
	// generate random blockStart
	blockStart := make([]byte, h.challengeDataSize)
	if _, err := h.cryptoRand.Read(blockStart); err != nil {
		return nil, err
	}

	// compute blockEnd
	blockEndInt := big.NewInt(0).SetBytes(blockStart)
	blockEndInt.Add(blockEndInt, h.challengeBlockSize)

	return &wow.Challenge{
		CostFunction: wow.Sha256,
		BlockStart:   blockStart,
		BlockEnd:     blockEndInt.Bytes(),
		Difficulty:   h.challengeDifficulty,
		ExpireAt:     h.now.Unix() + int64(h.challengeExpirationSec),
	}, nil
}

func (h *handler) validateChallenge(c *wow.Challenge) error {
	if c.ExpireAt <= h.now.Unix() {
		return errors.New("challenge has expired")
	}
	return nil
}

func (h *handler) signChallenge(c *wow.Challenge) ([]byte, error) {
	hash := hmac.New(sha256.New, h.serverSecret)

	// for answer to come from same IP address
	hash.Write([]byte(h.clientIP))

	if err := gob.NewEncoder(hash).Encode(c); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func (h *handler) validateSignature(c *wow.Challenge, providedSignature []byte) error {
	realSignature, err := h.signChallenge(c)
	if err != nil {
		return err
	}

	if !bytes.Equal(providedSignature, realSignature) {
		return errors.New("signature check failed")
	}

	return nil
}

func (h *handler) validateSolution(c *wow.Challenge, solution []byte) error {
	// check solution bounds
	solInt := big.NewInt(0).SetBytes(solution)
	startInt := big.NewInt(0).SetBytes(c.BlockStart)
	endInt := big.NewInt(0).SetBytes(c.BlockEnd)

	if solInt.Cmp(startInt) < 0 || solInt.Cmp(endInt) >= 0 {
		return errors.New("solution is out of challenge block")
	}

	// check solution correctness
	hash := sha256.New()
	hash.Write(solution)
	hashSum := hash.Sum(nil)
	hashSumInt := big.NewInt(0).SetBytes(hashSum)

	// works, because BitLen() won't count leading zero values
	if (hash.Size()*8)-hashSumInt.BitLen() < c.Difficulty {
		return fmt.Errorf("incorrect solution, hashsum: %x", hashSum)
	}

	return nil
}
