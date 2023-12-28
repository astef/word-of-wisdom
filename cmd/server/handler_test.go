package main

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/astef/word-of-wisdom/internal/log"
)

func Test_handler_handle(t *testing.T) {
	type fields struct {
		logger                  log.Logger
		now                     time.Time
		serverSecret            []byte
		clientIP                string
		challengeExpirationSec  int
		challengeDataSize       int
		challengeDifficulty     int
		challengeAvgSolutionNum int
		challengeBlockSize      *big.Int
	}
	type args struct {
		rq any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				logger:                  tt.fields.logger,
				now:                     tt.fields.now,
				serverSecret:            tt.fields.serverSecret,
				clientIP:                tt.fields.clientIP,
				challengeExpirationSec:  tt.fields.challengeExpirationSec,
				challengeDataSize:       tt.fields.challengeDataSize,
				challengeDifficulty:     tt.fields.challengeDifficulty,
				challengeAvgSolutionNum: tt.fields.challengeAvgSolutionNum,
				challengeBlockSize:      tt.fields.challengeBlockSize,
			}
			got, err := h.handle(tt.args.rq)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler.handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handler.handle() = %v, want %v", got, tt.want)
			}
		})
	}
}
