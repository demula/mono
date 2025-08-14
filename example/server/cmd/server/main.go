package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/demula/mono/core"
	"github.com/demula/mono/example/api"
)

func sayIt(w http.ResponseWriter, r *http.Request) {
	var it api.Hello
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&it)
	if err != nil {
		slog.Debug("failed to decode request", slog.String("error", err.Error()))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	slog.Info("got request", slog.String("who", it.Who))
	resp := &api.HelloResponse{
		Greeting: core.Say(it),
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "Failed to produce response", http.StatusInternalServerError)
		return
	}
	slog.Debug("sent response", slog.String("greeting", resp.Greeting))
}

func main() {
	slog.Info("starting server")
	mux := http.NewServeMux()
	mux.HandleFunc("/", sayIt)
	err := http.ListenAndServe(":8888", mux)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Info("server failed",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}