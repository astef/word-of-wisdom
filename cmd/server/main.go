package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/astef/word-of-wisdom/internal/api"
	"github.com/astef/word-of-wisdom/internal/log"
)

func init() {
	api.RegisterContracts()
}

func main() {
	// logging
	logInfo := log.NewInfo("")
	logWarn := log.NewWarn("")
	logErr := log.NewError("")

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
		s := <-exit
		logInfo.Println("received shutdown signal: ", s)
		cancel()
	}()

	// configuration
	cfg := getConfig()

	// pool for reusing buffers among connections
	bp := NewBufferPool(cfg.ConnectionReadBufferSize)

	// start tcp listener
	addr, err := net.ResolveTCPAddr("tcp", cfg.Address)
	if err != nil {
		logErr.Panicln(err.Error())
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logErr.Panicln(err.Error())
	}
	defer tcpListener.Close()

	for {
		// check shutdown
		if _, done := <-ctx.Done(); done {
			logInfo.Println("shutting down")
			break
		}

		// accept next connection
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			logWarn.Println("error accepting connection", err.Error())
			continue
		}

		go handleConnection(cfg, bp, ctx, tcpConn)
	}

	logInfo.Println("server exited gracefully")
}

func handleConnection(cfg *config, bp *bufferPool, ctx context.Context, tcpConn *net.TCPConn) {
	// connection logging
	clientAddr := tcpConn.RemoteAddr().String()
	logDebug := log.NewInfo(clientAddr)
	logInfo := log.NewInfo(clientAddr)
	logWarn := log.NewWarn(clientAddr)
	logErr := log.NewError(clientAddr)

	// recover from panics
	defer func() {
		if r := recover(); r != nil {
			logErr.Println("recovered from panic", r)
		}
	}()

	// borrow & release connection buffers
	buffers := bp.Borrow()
	defer func() {
		buffers.Release()
	}()

	// close connection
	defer func() {
		if err := tcpConn.Close(); err != nil {
			logWarn.Println("error closing connection", err.Error())
		}
	}()

	// configure connection
	now := time.Now()
	if err := tcpConn.SetDeadline(now.Add(time.Duration(cfg.ConnectionTimeoutMs) * time.Millisecond)); err != nil {
		logWarn.Println("error setting connection timeout", err.Error())
		return
	}
	if err := tcpConn.SetReadBuffer(cfg.ConnectionReadBufferSize); err != nil {
		logWarn.Println("error setting connection buffer size", err.Error())
		return
	}

	// "request-response" communication
	n, err := tcpConn.Read(buffers.ReadBuffer)
	if err != nil {
		logInfo.Println("error reading from connection", err.Error())
		return
	}
	requestBytes := buffers.ReadBuffer[:n]
	logDebug.Printf("request bytes (%d) %X\n", len(requestBytes), requestBytes)

	// decode request
	decoder := gob.NewDecoder(bytes.NewReader(requestBytes))
	var rq any
	if err := decoder.Decode(&rq); err != nil {
		logInfo.Println("error decoding request", err.Error())
		return
	}

	// invoke handler
	h := &handler{
		logDebug: logDebug,
		logInfo:  logInfo,
		logWarn:  logWarn,
		logErr:   logErr,
	}
	rs, err := h.handle(ctx, rq)
	if err != nil {
		logInfo.Println("error handling the request", err.Error())
		return
	}

	// encode response
	if err := gob.NewEncoder(buffers.WriteBuffer).Encode(rs); err != nil {
		logErr.Println("error encoding the response", err.Error())
		return
	}

	// send response
	rsBytes := buffers.WriteBuffer.Bytes()
	if _, err := tcpConn.Write(rsBytes); err != nil {
		logInfo.Println("error writing to connection", err.Error())
		return
	}
	logDebug.Printf("response bytes (%d) %X\n", len(rsBytes), rsBytes)
	logDebug.Println("elapsed:", time.Since(now).String())
}
