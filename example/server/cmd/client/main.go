package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/demula/mono/example/api"
)

const (
	url  = "localhost"
	port = 8888
)

func main() {
	if len(os.Args) != 3 {
		slog.Error("You must use 'go run cmd/client {{who}}' to call the server",
			slog.String("got", strings.Join(os.Args, " ")),
		)
	}

	hello := api.Hello{
		Who: os.Args[2],
	}
	body, err := json.Marshal(&hello)
	if err != nil {
		slog.Error("Failed to marshal JSON: %s",
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
		slog.Error("Failed to call localhost service: %s",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer resp.Body.Close()

	helloResp := &api.HelloResponse{}
	err = json.NewDecoder(resp.Body).Decode(helloResp)
	if err != nil {
		slog.Error("Failed to unmarshal JSON: %s",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	slog.Info("Got server response",
		slog.String("greeting", helloResp.Greeting),
	)
}