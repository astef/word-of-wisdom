package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	cfgutil "github.com/astef/word-of-wisdom/internal/config"
)

type config struct {
	Address                  string
	ConnectionTimeoutMs      int
	ConnectionReadBufferSize int
	ChallengeExpirationSec   int
	ChallengeDataSize        int
	ChallengeDifficulty      int
	ChallengeAvgSolutionNum  int
	ChallengeBlockSize       *big.Int
	ServerSecret             []byte
}

func getConfig() *config {
	cfg := &config{
		Address:                  cfgutil.ReadStrWithDefault("WOW_ADDRESS", ":5000"),
		ConnectionTimeoutMs:      cfgutil.MustReadIntWithDefault("WOW_CONN_TIMEOUT", 1000),
		ConnectionReadBufferSize: cfgutil.MustReadIntWithDefault("WOW_CONN_READ_BUFFER_SIZE", 64*1024), // 64KB
		ChallengeExpirationSec:   cfgutil.MustReadIntWithDefault("WOW_CHALLENGE_EXPIRATION_SEC", 3600), // 60 minutes
		ChallengeDataSize: cfgutil.MustReadIntWithDefault(
			"WOW_CHALLENGE_DATA_SIZE",
			300,
		), // 300 bytes, so all solutions exist between [0; 256^300)
		ChallengeDifficulty: cfgutil.MustReadIntWithDefault(
			"WOW_CHALLENGE_DIFFICULTY",
			25,
		), // 25 bits, so the chance of solution is 1 / 2^25
		ChallengeAvgSolutionNum: cfgutil.MustReadIntWithDefault(
			"WOW_CHALLENGE_AVG_SOLUTION_NUM",
			30,
		), // on average, 30 solutions should exist on the data range allocated to client
	}

	// practical protection against configuration mistakes, using reasonable boundaries
	// feel free to adjust them if you know what you do
	checkBounds("ConnectionTimeoutMs", cfg.ConnectionTimeoutMs, 100, 60000)
	checkBounds("ConnectionReadBufferSize", cfg.ConnectionReadBufferSize, 32*1024, 1024*1024)
	checkBounds("ChallengeExpirationSec", cfg.ChallengeExpirationSec, 10, 60*60*24*10)
	checkBounds("ChallengeDataSize", cfg.ChallengeDataSize, 100, 5000)
	checkBounds("ChallengeDifficulty", cfg.ChallengeDifficulty, 15, 80)
	checkBounds("ChallengeAvgSolutionNum", cfg.ChallengeAvgSolutionNum, 1, 100)

	// computed values
	cbs := big.NewInt(2)
	cbs.Exp(cbs, big.NewInt(int64(cfg.ChallengeDifficulty)), nil)
	cbs.Mul(cbs, big.NewInt(int64(cfg.ChallengeAvgSolutionNum)))
	cfg.ChallengeBlockSize = cbs

	cfg.ServerSecret = make([]byte, 512)
	if _, err := rand.Read(cfg.ServerSecret); err != nil {
		panic(err)
	}
	return cfg
}

func checkBounds(name string, value, validFrom, validTo int) {
	if value < validFrom || value >= validTo {
		panic(fmt.Sprintf("expected %s to be in range [%d; %d), got %d", name, validFrom, validTo, value))
	}
}
