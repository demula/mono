package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/demula/mono-example/api"
)

const (
	url  = "http://localhost"
	port = 8888
)

func main() {
	if len(os.Args) != 2 {
		slog.Error("you must use 'go run cmd/client/main.go {{who}}' or '{{executable}} {{who}}' to call the server",
			slog.String("got", strings.Join(os.Args, " ")),
		)
		os.Exit(1)
	}

	hello := api.Hello{
		Who: os.Args[1],
	}
	body, err := json.Marshal(&hello)
	if err != nil {
		slog.Error("failed to marshal JSON: %s",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s:%d", url, port),
		"application/json; charset=UTF-8",
		bytes.NewBuffer(body),
	)
	if err != nil {
		slog.Error("failed to call localhost service: %s",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer resp.Body.Close()

	helloResp := &api.HelloResponse{}
	err = json.NewDecoder(resp.Body).Decode(helloResp)
	if err != nil {
		slog.Error("failed to unmarshal JSON: %s",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	slog.Info("got server response",
		slog.String("greeting", helloResp.Greeting),
	)
}