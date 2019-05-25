package main

import (
	"context"
	"flag"
	"time"

	rpc "github.com/onlinecity/ocmg-lib/src/go/pkg/rpc"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
)

func main() {
	var endpoint = flag.String("endpoint", "tcp://localhost:7200", "where to connect")
	var retryrate = flag.Uint("retryrate", 1000, "how often to retry in ms")
	var timeout = flag.Uint("timeout", 2000, "timeout in ms, total for all attempts")
	var development = flag.Bool("dev", false, "run in development mode")
	flag.Parse()

	var logger *zap.Logger
	if *development {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // nolint:errcheck

	zap.S().Infof("connecting to %q\n", *endpoint)
	client, err := rpc.NewConnection(zmq.REQ)
	if err != nil {
		zap.S().Fatal(err)
	}
	defer client.Close()
	if err := client.Connect(*endpoint); err != nil {
		zap.S().Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(*timeout)*time.Millisecond)
	defer cancel()

	rate := time.Duration(*retryrate) * time.Millisecond

	if body, err := client.CallRepeat("healthz", rate, ctx); err != nil || body != 0 {
		if err != nil && err == context.DeadlineExceeded {
			zap.S().Fatal("timed out")
		} else {
			zap.S().Fatalw("healthz failed", "err", err, "body", body)
		}
	}
}
