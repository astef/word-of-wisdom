package main

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	mathrand "math/rand"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/astef/word-of-wisdom/internal/log"
)

func main() {
	// logging
	logger := log.NewDefaultLogger()

	// configuration
	cfg := getConfig()

	// start tcp listener
	addr, err := net.ResolveTCPAddr("tcp", cfg.Address)
	if err != nil {
		logger.Error().Println(err.Error())
		os.Exit(1)
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Error().Println(err.Error())
		os.Exit(1)
	}
	defer tcpListener.Close()

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
		s := <-exit
		logger.Info().Println("received shutdown signal: ", s)
		cancel()
		// tcpListener.AcceptTCP() may hang when not accepting connections, so forcibly close
		tcpListener.Close()
	}()

	logger.Info().Println("listening on", cfg.Address)

	for {
		// check shutdown
		if err := ctx.Err(); err != nil {
			logger.Info().Println("shutting down, because:", err.Error())
			break
		}

		// accept next connection
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			logger.Warn().Println("error accepting connection", err.Error())
			continue
		}

		go handleConnection(cfg, tcpConn)
	}

	logger.Info().Println("server exited gracefully")
}

func handleConnection(cfg *config, tcpConn *net.TCPConn) {
	// connection logging
	clientAddr := tcpConn.RemoteAddr().String()
	logger := log.NewDefaultLogger().Prefix(clientAddr)

	// recover from panics
	defer func() {
		if r := recover(); r != nil {
			logger.Error().Println("recovered from panic:", r, "\n", string(debug.Stack()))
		}
	}()

	// close connection
	defer func() {
		if err := tcpConn.Close(); err != nil {
			logger.Warn().Println("error closing connection", err.Error())
		}
	}()

	// configure connection
	if err := configureTCPConn(cfg, tcpConn, logger); err != nil {
		return
	}

	// "request-response" communication
	// decode request
	decoder := gob.NewDecoder(tcpConn)
	var rq any
	if err := decoder.Decode(&rq); err != nil {
		logger.Info().Println("error decoding request", err.Error())
		return
	}

	// invoke handler
	clientIP, _, _ := strings.Cut(clientAddr, ":")
	h := &handler{
		logger:   logger,
		clientIP: clientIP,
		// important to have fresh time here
		now:                     time.Now(),
		serverSecret:            cfg.ServerSecret,
		challengeExpirationSec:  cfg.ChallengeExpirationSec,
		challengeDataSize:       cfg.ChallengeDataSize,
		challengeDifficulty:     cfg.ChallengeDifficulty,
		challengeAvgSolutionNum: cfg.ChallengeAvgSolutionNum,
		challengeBlockSize:      cfg.ChallengeBlockSize,
		cryptoRand:              rand.Reader,
		quoteRandIntn:           mathrand.Intn,
	}
	rs, err := h.handle(rq)
	if err != nil {
		logger.Info().Println("error handling the request", err.Error())
		return
	}

	// encode response
	if err := gob.NewEncoder(tcpConn).Encode(rs); err != nil {
		logger.Info().Println("error encoding the response", err.Error())
		return
	}
}

func configureTCPConn(cfg *config, tcpConn *net.TCPConn, logger log.Logger) error {
	if err := tcpConn.SetDeadline(time.Now().Add(time.Duration(cfg.ConnectionTimeoutMs) * time.Millisecond)); err != nil {
		logger.Warn().Println("error setting connection timeout", err.Error())
		return err
	}
	if err := tcpConn.SetReadBuffer(cfg.ConnectionReadBufferSize); err != nil {
		logger.Warn().Println("error setting connection buffer size", err.Error())
		return err
	}
	return nil
}
