package api

import (
	"net/http"
	"strings"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", onlyMethod(http.MethodPost, handleLogin))
	mux.HandleFunc("/logout", onlyMethod(http.MethodPost, auth(handleLogout)))

	mux.HandleFunc("/containers", onlyMethod(http.MethodGet, auth(handleListContainers)))
	mux.HandleFunc("/containers/", auth(routeContainerByID))

	return cors(mux)
}

func routeContainerByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/start") && r.Method == http.MethodPost:
		handleStartContainer(w, r)
	case strings.HasSuffix(path, "/stop") && r.Method == http.MethodPost:
		handleStopContainer(w, r)
	case strings.HasSuffix(path, "/restart") && r.Method == http.MethodPost:
		handleRestartContainer(w, r)
	case strings.HasSuffix(path, "/logs/stream") && r.Method == http.MethodGet:
		handleContainerLogsStream(w, r)
	case strings.HasSuffix(path, "/logs") && r.Method == http.MethodGet:
		handleContainerLogs(w, r)
	case strings.HasSuffix(path, "/stats") && r.Method == http.MethodGet:
		handleContainerStats(w, r)
	case strings.HasSuffix(path, "/inspect") && r.Method == http.MethodGet:
		handleContainerInspect(w, r)
	default:
		writeError(w, http.StatusNotFound, "route not found")
	}
}

func onlyMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}
