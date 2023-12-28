package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

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
		Address:             cfgutil.ReadStrWithDefault("WOW_ADDRESS", ":5000"),
		ConnectionTimeoutMs: cfgutil.ReadIntWithDefault("WOW_CONN_TIMEOUT", 1000, onConfigFailure),
		ConnectionReadBufferSize: cfgutil.ReadIntWithDefault(
			"WOW_CONN_READ_BUFFER_SIZE",
			64*1024,
			onConfigFailure,
		), // 64KB
		ChallengeExpirationSec: cfgutil.ReadIntWithDefault(
			"WOW_CHALLENGE_EXPIRATION_SEC",
			3600,
			onConfigFailure,
		), // 60 minutes
		ChallengeDataSize: cfgutil.ReadIntWithDefault(
			"WOW_CHALLENGE_DATA_SIZE",
			300,
			onConfigFailure,
		), // 300 bytes, so all solutions exist between [0; 256^300)
		ChallengeDifficulty: cfgutil.ReadIntWithDefault(
			"WOW_CHALLENGE_DIFFICULTY",
			20,
			onConfigFailure,
		), // 20 bits, so the chance of solution is 1 / 2^20
		ChallengeAvgSolutionNum: cfgutil.ReadIntWithDefault(
			"WOW_CHALLENGE_AVG_SOLUTION_NUM",
			30,
			onConfigFailure,
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
	cfg.ChallengeBlockSize = challengeBlockSize(cfg.ChallengeDifficulty, cfg.ChallengeAvgSolutionNum)

	// TODO: this is problematic in the distributed server scenario, this key should be shared among different servers,
	// but for demo we'll just generate it at startup
	cfg.ServerSecret = make([]byte, 512)
	if _, err := rand.Read(cfg.ServerSecret); err != nil {
		print("failed generating random data")
		os.Exit(1)
	}
	return cfg
}

func challengeBlockSize(difficulty int, avgSolutionNum int) *big.Int {
	cbs := big.NewInt(2)
	cbs.Exp(cbs, big.NewInt(int64(difficulty)), nil)
	cbs.Mul(cbs, big.NewInt(int64(avgSolutionNum)))
	return cbs
}

func checkBounds(name string, value, validFrom, validTo int) {
	if value < validFrom || value >= validTo {
		print(fmt.Sprintf("expected %s to be in range [%d; %d), got %d", name, validFrom, validTo, value))
		os.Exit(1)
	}
}

func onConfigFailure(err error) {
	print(err.Error())
	os.Exit(1)
}
