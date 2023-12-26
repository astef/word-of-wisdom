package main

import (
	"context"
	"encoding/gob"
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
		panic(err)
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Error().Println(err.Error())
		panic(err)
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

		go handleConnection(cfg, ctx, tcpConn)
	}

	logger.Info().Println("server exited gracefully")
}

func handleConnection(cfg *config, ctx context.Context, tcpConn *net.TCPConn) {
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
	now := time.Now()
	if err := tcpConn.SetDeadline(now.Add(time.Duration(cfg.ConnectionTimeoutMs) * time.Millisecond)); err != nil {
		logger.Warn().Println("error setting connection timeout", err.Error())
		return
	}
	if err := tcpConn.SetReadBuffer(cfg.ConnectionReadBufferSize); err != nil {
		logger.Warn().Println("error setting connection buffer size", err.Error())
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
	clientIp, _, _ := strings.Cut(clientAddr, ":")
	h := &handler{
		logger:   logger,
		clientIP: clientIp,
		// important to have fresh time here
		now:                     time.Now(),
		serverSecret:            cfg.ServerSecret,
		challengeExpirationSec:  cfg.ChallengeExpirationSec,
		challengeDataSize:       cfg.ChallengeDataSize,
		challengeDifficulty:     cfg.ChallengeDifficulty,
		challengeAvgSolutionNum: cfg.ChallengeAvgSolutionNum,
		challengeBlockSize:      cfg.ChallengeBlockSize,
	}
	rs, err := h.handle(ctx, rq)
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
