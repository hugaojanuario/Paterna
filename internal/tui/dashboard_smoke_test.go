package tui

import (
	"strings"
	"testing"

	"github.com/hugaojanuario/Paterna/internal/system"
)

// Smoke: a View do dashboard não pode entrar em panic com larguras variadas,
// inclusive terminal pequeno e snapshot vazio.
func TestDashboardViewNoPanic(t *testing.T) {
	snap := system.Snapshot{
		Hostname:   "test",
		OS:         "linux 1.0",
		CPUPerCore: []float64{12, 80, 5, 99, 40, 0},
		CPUTotal:   42,
		CPUCount:   6,
		MemUsed:    4 << 30,
		MemTotal:   16 << 30,
		MemPercent: 25,
		Disks:      []system.DiskUsage{{Path: "/", Used: 50 << 30, Total: 100 << 30, UsedPercent: 50}},
		Procs:      []system.ProcInfo{{PID: 1, Name: "init", CPU: 1.2, MemMB: 10}},
	}

	for _, w := range []int{0, 20, 40, 80, 120, 200} {
		for _, h := range []int{0, 10, 24, 50} {
			m := NewDashboardModel()
			m.width, m.height = w, h
			updated, _ := m.Update(dashboardDataMsg{snap: snap, containers: nil, dockerErr: nil})
			out := updated.View()
			if strings.Contains(out, "panic") {
				t.Fatalf("saída inesperada em %dx%d", w, h)
			}
		}
	}
}

func TestCollectorNoPanic(t *testing.T) {
	c := system.NewCollector()
	_ = c.Collect()
	snap := c.Collect()
	if snap.CPUCount == 0 && len(snap.CPUPerCore) == 0 {
		t.Log("aviso: sem dados de CPU neste ambiente")
	}
}
