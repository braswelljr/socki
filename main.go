package main

import (
	"context"
	"flag"
	"net"

	"github.com/braswelljr/socki/utils"
	"github.com/braswelljr/socki/www"
)

var serverAddr = flag.String("addr", ":5000", "server address of the api gateway and web app")

func main() {
	// parse flags
	flag.Parse()
	// create logger
	logger := utils.NewLogger()

	listener, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		logger.Fatal().Err(err).Str("address", *serverAddr).Msg("failed to listen to address")
	}

	logger.Info().Str("address", *serverAddr).Msg("listening on address")

	s := NewServer(logger, listener, www.Assets)

	ctx := context.Background()
	err = s.Run(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to run server")
	}

	defer func(server *Server) {
		err := server.Close()
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to close server")
		}
	}(s)
}
