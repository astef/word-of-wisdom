package main

import (
	"context"
	"crypto/sha256"
	"math/big"
	"math/bits"
	"os"

	"github.com/astef/word-of-wisdom/internal/log"
	"github.com/astef/word-of-wisdom/internal/wow"
)

func main() {
	logger := log.NewDefaultLogger()

	ctx := context.Background()

	cfg := getConfig()

	client := wow.NewClient(&wow.ClientConfig{Address: cfg.Address})

	for i := 0; i < cfg.QuotesNum; i++ {
		quote, err := getQuote(ctx, client, logger)
		if err != nil {
			logger.Error().Printf("Error getting a quote: %s", err.Error())
			os.Exit(1)
		}
		if quote == "" {
			logger.Info().Println("No solution, trying another challenge")
			i--
			continue
		}

		logger.Info().Println("Awarded with a quote:", quote)
	}

}

func getQuote(ctx context.Context, client wow.Client, logger log.Logger) (string, error) {
	chResp, err := client.GetChallenge(ctx, &wow.ChallengeRequest{})
	if err != nil {
		logger.Error().Printf("Error getting challenge: %s", err.Error())
		os.Exit(1)
	}
	logger.Info().Printf("Got challenge, start solving.")

	solution, hashSum := findSolution(
		chResp.Challenge.BlockStart,
		chResp.Challenge.BlockEnd,
		chResp.Challenge.Difficulty,
	)
	if solution == nil {
		return "", nil
	}

	logger.Info().Printf("Found solution: %x with checksum: %x", solution, hashSum)
	quoteResp, err := client.GetQuote(ctx, &wow.QuoteRequest{
		ChallengeResponse: *chResp,
		Solution:          solution,
	})

	if err != nil {
		return "", err
	}

	return quoteResp.Quote, nil
}

func findSolution(start, end []byte, difficulty int) ([]byte, []byte) {
	startInt := big.NewInt(0).SetBytes(start)
	endInt := big.NewInt(0).SetBytes(end)
	incr := big.NewInt(1)
	for current := startInt; current.Cmp(endInt) < 0; current.Add(current, incr) {
		hash := sha256.New()
		hash.Write(current.Bytes())
		hashSum := hash.Sum(nil)

		if leadingZeroBits(hashSum) >= difficulty {
			return current.Bytes(), hashSum
		}
	}
	return nil, nil
}

func leadingZeroBits(b []byte) int {
	totalLeadingZeros := 0
	for i := 0; i < len(b); i++ {
		leadingZeros := bits.LeadingZeros8(b[i])
		totalLeadingZeros += leadingZeros
		if leadingZeros != 8 {
			break
		}
	}
	return totalLeadingZeros
}
