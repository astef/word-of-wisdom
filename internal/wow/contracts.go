package wow

import "encoding/gob"

type CostFunction = int

const (
	Sha256 CostFunction = iota
)

type ChallengeRequest struct {
}

type Challenge struct {
	// What function to be used to search for an answer
	CostFunction CostFunction
	// Starting number from where the search for the solution should start (inclusive)
	// Big-endian unsigned integer
	BlockStart []byte
	// End number where the search for the solution should stop (exclusive)
	// Big-endian unsigned integer
	BlockEnd []byte
	// Required number of zero bits in the beginning of CostFunction output
	Difficulty int
	// time in Unix seconds, when the challenge expires
	ExpireAt int64
}

type ChallengeResponse struct {
	// Challenge, which should be solved
	Challenge *Challenge
	// Challenge signature, which should be passed as-is with further requests
	Signature []byte
}

type QuoteRequest struct {
	ChallengeResponse

	// Solution of the challenge on the range starting from BlockStart to BlockEnd
	// CostFunction applied on the Solution should return the result, which has number of leading zero bytes equal to
	// Difficulty.
	// Big-endian unsigned integer
	Solution []byte
}

type QuoteResponse struct {
	Quote string
}

func init() {
	// required for encoding/decoding to/from interface work
	gob.Register(&ChallengeRequest{})
	gob.Register(&ChallengeResponse{})
	gob.Register(&QuoteRequest{})
	gob.Register(&QuoteResponse{})
}
