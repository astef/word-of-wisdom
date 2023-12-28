package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/astef/word-of-wisdom/internal/log"
	"github.com/astef/word-of-wisdom/internal/wow"
)

type testCase struct {
	name          string
	handler       testHandler
	rq            any
	responseCheck func(t *testing.T, tt *testCase, got any)
	wantErr       string
}

func Test_handler_handle(t *testing.T) {

	testHandler := newTestHandler()

	correctSolution := decodeHex(
		"2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2C3C",
	)

	tests := []*testCase{
		{
			name:          "ChallengeRequest - positive",
			handler:       testHandler,
			rq:            &wow.ChallengeRequest{},
			responseCheck: checkChallengeResponse,
		},
		{
			name:    "QuoteRequest - positive",
			handler: testHandler,
			rq: &wow.QuoteRequest{
				ChallengeResponse: *testHandler.expectedChallengeResponse,
				Solution:          correctSolution,
			},
			responseCheck: checkQuoteResponse,
		},
		{
			name:    "QuoteRequest - negative - solution out of block",
			handler: testHandler,
			rq: &wow.QuoteRequest{
				ChallengeResponse: *testHandler.expectedChallengeResponse,
				Solution:          testHandler.expectedChallengeResponse.Challenge.BlockEnd,
			},
			wantErr: "solution is out of challenge block",
		},
		{
			name:    "QuoteRequest - negative - solution not correct",
			handler: testHandler,
			rq: &wow.QuoteRequest{
				ChallengeResponse: *testHandler.expectedChallengeResponse,
				Solution: decodeHex(
					"2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2C3B",
				),
			},
			wantErr: "incorrect solution",
		},
		{
			name:    "QuoteRequest - negative - solution expired",
			handler: testHandler.MakeExpired(),
			rq: &wow.QuoteRequest{
				ChallengeResponse: *testHandler.expectedChallengeResponse,
				Solution:          correctSolution,
			},
			wantErr: "challenge has expired",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.handler.handle(tt.rq)

			if tt.wantErr != "" {
				errText := ""
				if err != nil {
					errText = err.Error()
				}
				if !strings.HasPrefix(errText, tt.wantErr) {
					t.Errorf("error=%v, wantErr=%v", errText, tt.wantErr)
					return
				}
			}

			if tt.responseCheck != nil {
				tt.responseCheck(t, tt, got)
			}
		})
	}
}

type testHandler struct {
	handler
	expectedChallengeResponse *wow.ChallengeResponse
}

func (th testHandler) MakeExpired() testHandler {
	h := th.handler
	h.now = h.now.Add(time.Second * time.Duration(h.challengeExpirationSec+1))
	return testHandler{
		handler:                   h,
		expectedChallengeResponse: th.expectedChallengeResponse,
	}
}

func newTestHandler() testHandler {
	difficulty := 10
	avgSolutionNum := 1
	challengeDataSize := 40
	blockStart := bytes.Repeat([]byte{42}, challengeDataSize)
	now := time.Date(2100, time.August, 1, 2, 3, 4, 5, time.UTC)
	challengeExpirationSec := 60

	return testHandler{
		handler: handler{
			logger:                  log.NewDiscardLogger(),
			now:                     now,
			serverSecret:            []byte{1, 2, 3, 4, 5},
			clientIP:                "1.2.3.4",
			challengeExpirationSec:  challengeExpirationSec,
			challengeDataSize:       challengeDataSize,
			challengeDifficulty:     difficulty,
			challengeAvgSolutionNum: avgSolutionNum,
			challengeBlockSize:      challengeBlockSize(difficulty, avgSolutionNum),
			cryptoRand:              NewTestRandReader(challengeDataSize, blockStart),
			quoteRandIntn:           func(n int) int { return 13 },
		},
		expectedChallengeResponse: &wow.ChallengeResponse{
			Challenge: &wow.Challenge{
				CostFunction: wow.Sha256,
				BlockStart:   blockStart,
				BlockEnd: decodeHex(
					"2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A36AA2A2A",
				),
				Difficulty: difficulty,
				ExpireAt:   now.Unix() + int64(challengeExpirationSec),
			},
			Signature: decodeHex("7AFF85394620E722065167696A7CD4AE6E712614AFB00DFF165F340EDF22475D"),
		},
	}
}

type testCryptoRandReader struct {
	expectSize  int
	returnBytes []byte
}

func (t *testCryptoRandReader) Read(p []byte) (n int, err error) {
	if len(p) != t.expectSize {
		return 0, fmt.Errorf("unexpected requested buffer size %d, want %d", len(p), t.expectSize)
	}

	return copy(p, t.returnBytes), nil
}

func NewTestRandReader(expectSize int, returnBytes []byte) io.Reader {
	return &testCryptoRandReader{
		expectSize:  expectSize,
		returnBytes: returnBytes,
	}
}

func checkChallengeResponse(t *testing.T, tt *testCase, got any) {
	resp := got.(*wow.ChallengeResponse)

	if len(resp.Signature) < sha256.Size {
		t.Errorf("Signature is %d, want %d", len(resp.Signature), sha256.Size)
	}

	if resp.Challenge == nil {
		t.Error("Challenge is nil")
	}

	if resp.Challenge.CostFunction != wow.Sha256 {
		t.Errorf("CostFunction is %d, want %d", resp.Challenge.CostFunction, wow.Sha256)
	}

	blockStartInt := big.NewInt(0).SetBytes(resp.Challenge.BlockStart)
	blockEndInt := big.NewInt(0).SetBytes(resp.Challenge.BlockEnd)
	if blockEndInt.Cmp(blockStartInt) <= 0 {
		t.Errorf("expected BlockEnd to be greater then BlockStart")
	}

	rangeInt := big.NewInt(0)
	rangeInt.Sub(blockEndInt, blockStartInt)
	if tt.handler.challengeBlockSize.Cmp(rangeInt) != 0 {
		t.Errorf(
			"expected BlockEnd-BlockStart to be %s, got %s",
			tt.handler.challengeBlockSize.String(),
			rangeInt.String(),
		)
	}

	if resp.Challenge.Difficulty != tt.handler.challengeDifficulty {
		t.Errorf("Difficulty is %d, expected %d", resp.Challenge.Difficulty, tt.handler.challengeDifficulty)
	}

	if resp.Challenge.ExpireAt-tt.handler.now.Unix() != int64(tt.handler.challengeExpirationSec) {
		t.Errorf("ExpireAt is %d, expected %d", resp.Challenge.ExpireAt, int64(tt.handler.challengeExpirationSec))
	}
}

func checkQuoteResponse(t *testing.T, tt *testCase, got any) {
	resp := got.(*wow.QuoteResponse)
	expectedQuote := "Opportunities don't happen, you create them. ~Chris Grosser"
	if resp.Quote != expectedQuote {
		t.Errorf("expected quote to be '%s', got '%s'", expectedQuote, resp.Quote)
	}
}
func decodeHex(s string) []byte {
	res, _ := hex.DecodeString(s)
	return res
}
