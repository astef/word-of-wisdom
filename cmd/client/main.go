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
		if quote != "" {
			logger.Info().Println("Awarded with a quote:", quote)
		}
	}

}

func getQuote(ctx context.Context, client wow.Client, logger log.Logger) (string, error) {
	chResp, err := client.GetChallenge(ctx, &wow.ChallengeRequest{})
	if err != nil {
		logger.Error().Printf("Error getting challenge: %s", err.Error())
		os.Exit(1)
	}
	logger.Info().Printf("Got challenge, start solving.")

	current := big.NewInt(0).SetBytes(chResp.Challenge.BlockStart)
	end := big.NewInt(0).SetBytes(chResp.Challenge.BlockEnd)

	for {
		if current.Cmp(end) >= 0 {
			logger.Warn().
				Println("Reached the end of block without finding any value. Unlucky, or server is too restrictive.")
			return "", err
		}

		hash := sha256.New()
		hash.Write(current.Bytes())
		hashSum := hash.Sum(nil)

		if leadingZeroBits(hashSum) < chResp.Challenge.Difficulty {
			current.Add(current, big.NewInt(1))
			continue
		}

		logger.Info().Printf("Found solution: %x with checksum: %x", current.Bytes(), hashSum)

		quoteResp, err := client.GetQuote(ctx, &wow.QuoteRequest{
			ChallengeResponse: *chResp,
			Solution:          current.Bytes(),
		})

		if err != nil {
			return "", err
		}

		return quoteResp.Quote, nil
	}
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
