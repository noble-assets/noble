package jester

import (
	"net/http"
	"strings"

	jester "jester.noble.xyz/api"
)

func NewJesterGRPCClient(grpcAddr string) (gRPCclient jester.QueryServiceClient) {
	if !strings.Contains(grpcAddr, "://") {
		grpcAddr = "http://" + grpcAddr
	}

	return jester.NewQueryServiceClient(
		http.DefaultClient,
		grpcAddr,
	)
}
