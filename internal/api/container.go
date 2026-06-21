package api

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/hugaojanuario/Paterna/internal/container"
	errorsx "github.com/hugaojanuario/Paterna/pkg/errors"
)

func handleListContainers(w http.ResponseWriter, r *http.Request) {
	rows, err := container.List(true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []container.ContainerInfo{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func handleStartContainer(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/start")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	if err := container.StartContainer(id); err != nil {
		writeContainerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "started", "id": id})
}

func handleStopContainer(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/stop")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	if err := container.StopContainer(id); err != nil {
		writeContainerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped", "id": id})
}

func handleRestartContainer(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/restart")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	if err := container.RestartContainer(id); err != nil {
		writeContainerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted", "id": id})
}

func handleContainerLogsStream(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/logs/stream")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	reader, err := container.StreamContainerLogs(id)
	if err != nil {
		writeContainerError(w, err)
		return
	}
	defer reader.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		stdcopy.StdCopy(pw, pw, reader)
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
			flusher.Flush()
		}
	}
}

func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/logs")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	logs, err := container.GetContainerLogs(id)
	if err != nil {
		writeContainerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id, "logs": logs})
}

func handleContainerStats(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/stats")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	stats, err := container.GetContainerStats(id)
	if err != nil {
		writeContainerError(w, err)
		return
	}

	cpu, mem, memLimit := computeUsage(stats)

	writeJSON(w, http.StatusOK, map[string]any{
		"id":              id,
		"cpu_percent":     cpu,
		"memory_mb":       mem,
		"memory_limit_mb": memLimit,
	})
}

func handleContainerInspect(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/containers/", "/inspect")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing container id")
		return
	}

	state, err := container.InspectContainer(id)
	if err != nil {
		writeContainerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, state)
}

func pathID(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)

	if suffix != "" {
		if !strings.HasSuffix(rest, suffix) {
			return ""
		}
		rest = strings.TrimSuffix(rest, suffix)
	}

	return rest
}

func writeContainerError(w http.ResponseWriter, err error) {
	if errors.Is(err, errorsx.ErrNotFound) {
		writeError(w, http.StatusNotFound, "container not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func computeUsage(s container.ContainerStats) (float64, float64, float64) {
	cpuDelta := float64(s.CpuStats.CpuUsage.TotalUsage) - float64(s.PreCpuStats.CpuUsage.TotalUsage)
	sysDelta := float64(s.CpuStats.SystemCpuUsage) - float64(s.PreCpuStats.SystemCpuUsage)

	cpu := 0.0
	if sysDelta > 0 && cpuDelta > 0 && s.CpuStats.OnlineCpus > 0 {
		cpu = (cpuDelta / sysDelta) * float64(s.CpuStats.OnlineCpus) * 100.0
	}

	memMB := float64(s.MemoryStats.Usage) / 1024.0 / 1024.0
	memLimitMB := float64(s.MemoryStats.Limit) / 1024.0 / 1024.0

	return cpu, memMB, memLimitMB
}
