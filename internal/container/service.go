package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/hugaojanuario/Paterna/pkg/docker"
	errors "github.com/hugaojanuario/Paterna/pkg/errors"
)

type ContainerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

type ContainerStats struct {
	CpuStats    CpuStats    `json:"cpu_stats"`
	PreCpuStats CpuStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
}

type CpuStats struct {
	CpuUsage       CpuUsage `json:"cpu_usage"`
	SystemCpuUsage uint64   `json:"system_cpu_usage"`
	OnlineCpus     uint64   `json:"online_cpus"`
}

type CpuUsage struct {
	TotalUsage uint64 `json:"total_usage"`
}

type MemoryStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

type ContainerState struct {
	ID         string
	Name       string
	Status     string
	OOMKilled  bool
	ExitCode   int
	ExitError  string
	FinishedAt time.Time
}

func List(all bool) ([]ContainerInfo, error) {
	client, err := docker.GetClient()
	if err != nil {
		return nil, err
	}

	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		result = append(result, ContainerInfo{
			ID:     c.ID,
			Name:   name,
			Image:  c.Image,
			Status: c.Status,
		})
	}

	return result, nil
}

func StartContainer(id string) error {
	client, err := docker.GetClient()
	if err != nil {
		return err
	}

	err = client.ContainerStart(context.Background(), id, container.StartOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		return err
	}

	return nil
}

func StopContainer(id string) error {
	client, err := docker.GetClient()
	if err != nil {
		return err
	}

	err = client.ContainerStop(context.Background(), id, container.StopOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		return err
	}

	return nil
}

func RestartContainer(id string) error {
	client, err := docker.GetClient()
	if err != nil {
		return err
	}

	err = client.ContainerRestart(context.Background(), id, container.StopOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		return err
	}

	return nil
}

func GetContainerLogs(id string) (string, error) {
	client, err := docker.GetClient()
	if err != nil {
		return "", err
	}

	menu := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "50",
	}

	reader, err := client.ContainerLogs(context.Background(), id, menu)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return "", errors.ErrNotFound
		}
		return "", err
	}
	defer reader.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return "", err
	}

	return stdout.String() + stderr.String(), nil
}

func GetContainerStats(id string) (ContainerStats, error) {
	client, err := docker.GetClient()
	if err != nil {
		return ContainerStats{}, err
	}

	stats, err := client.ContainerStats(context.Background(), id, false)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ContainerStats{}, errors.ErrNotFound
		}
		return ContainerStats{}, err
	}
	defer stats.Body.Close()

	var containerStats ContainerStats
	err = json.NewDecoder(stats.Body).Decode(&containerStats)
	if err != nil {
		return ContainerStats{}, err
	}

	return containerStats, nil
}

func StreamContainerLogs(id string) (io.ReadCloser, error) {
	client, err := docker.GetClient()
	if err != nil {
		return nil, err
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50",
	}

	reader, err := client.ContainerLogs(context.Background(), id, options)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	return reader, nil
}

// ComputeUsage transforma estatísticas brutas do Docker em valores prontos
// para exibição: cpu em %, memória usada em MB e limite em MB.
func ComputeUsage(s ContainerStats) (cpu, memMB, memLimitMB float64) {
	cpuDelta := float64(s.CpuStats.CpuUsage.TotalUsage) - float64(s.PreCpuStats.CpuUsage.TotalUsage)
	sysDelta := float64(s.CpuStats.SystemCpuUsage) - float64(s.PreCpuStats.SystemCpuUsage)

	if sysDelta > 0 && cpuDelta > 0 && s.CpuStats.OnlineCpus > 0 {
		cpu = (cpuDelta / sysDelta) * float64(s.CpuStats.OnlineCpus) * 100.0
	}

	memMB = float64(s.MemoryStats.Usage) / 1024.0 / 1024.0
	memLimitMB = float64(s.MemoryStats.Limit) / 1024.0 / 1024.0
	return
}

func InspectContainer(id string) (ContainerState, error) {
	cl, err := docker.GetClient()
	if err != nil {
		return ContainerState{}, err
	}

	info, err := cl.ContainerInspect(context.Background(), id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ContainerState{}, errors.ErrNotFound
		}
		return ContainerState{}, err
	}

	name := strings.TrimPrefix(info.Name, "/")

	var finishedAt time.Time
	if t, err := time.Parse(time.RFC3339Nano, info.State.FinishedAt); err == nil {
		finishedAt = t
	}

	return ContainerState{
		ID:         info.ID,
		Name:       name,
		Status:     info.State.Status,
		OOMKilled:  info.State.OOMKilled,
		ExitCode:   info.State.ExitCode,
		ExitError:  info.State.Error,
		FinishedAt: finishedAt,
	}, nil
}
