// Package system coleta métricas da máquina local (CPU, memória, disco,
// rede e processos) usando gopsutil. É a fundação do dashboard estilo btop.
package system

import (
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Snapshot é uma fotografia das métricas num instante.
type Snapshot struct {
	Hostname string
	OS       string
	Uptime   time.Duration
	Load1    float64
	Load5    float64
	Load15   float64

	CPUPerCore []float64 // % por core
	CPUTotal   float64   // % média
	CPUCount   int

	MemUsed    uint64
	MemTotal   uint64
	MemPercent float64
	SwapUsed   uint64
	SwapTotal  uint64

	Disks []DiskUsage

	NetRecvRate float64 // bytes/s
	NetSentRate float64 // bytes/s

	Procs []ProcInfo
}

type DiskUsage struct {
	Path        string
	Used        uint64
	Total       uint64
	UsedPercent float64
}

type ProcInfo struct {
	PID   int32
	Name  string
	CPU   float64 // % instantâneo
	MemMB float64
}

// Collector mantém estado entre coletas para calcular taxas (rede) e o uso
// instantâneo de CPU por processo (delta de tempo de CPU).
type Collector struct {
	lastNetRecv uint64
	lastNetSent uint64
	lastNetTime time.Time

	lastProcCPU map[int32]procCPU
}

type procCPU struct {
	total float64 // user+system acumulado em segundos
	when  time.Time
}

// MaxProcs limita quantos processos aparecem na tabela.
const MaxProcs = 20

func NewCollector() *Collector {
	return &Collector{lastProcCPU: map[int32]procCPU{}}
}

// Collect monta um Snapshot. Erros parciais são tolerados: um subsistema
// indisponível (ex: load average no Windows) apenas zera aquele campo.
func (c *Collector) Collect() Snapshot {
	var s Snapshot

	if info, err := host.Info(); err == nil {
		s.Hostname = info.Hostname
		s.OS = info.OS + " " + info.PlatformVersion
		s.Uptime = time.Duration(info.Uptime) * time.Second
	}

	if avg, err := load.Avg(); err == nil {
		s.Load1, s.Load5, s.Load15 = avg.Load1, avg.Load5, avg.Load15
	}

	c.collectCPU(&s)
	c.collectMem(&s)
	c.collectDisks(&s)
	c.collectNet(&s)
	c.collectProcs(&s)

	return s
}

func (c *Collector) collectCPU(s *Snapshot) {
	if per, err := cpu.Percent(0, true); err == nil {
		s.CPUPerCore = per
		s.CPUCount = len(per)
	}
	if total, err := cpu.Percent(0, false); err == nil && len(total) > 0 {
		s.CPUTotal = total[0]
	}
}

func (c *Collector) collectMem(s *Snapshot) {
	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemUsed = vm.Used
		s.MemTotal = vm.Total
		s.MemPercent = vm.UsedPercent
	}
	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapUsed = sw.Used
		s.SwapTotal = sw.Total
	}
}

func (c *Collector) collectDisks(s *Snapshot) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return
	}

	seen := map[string]bool{}
	for _, p := range parts {
		if seen[p.Mountpoint] {
			continue
		}
		seen[p.Mountpoint] = true

		u, err := disk.Usage(p.Mountpoint)
		if err != nil || u.Total == 0 {
			continue
		}
		s.Disks = append(s.Disks, DiskUsage{
			Path:        p.Mountpoint,
			Used:        u.Used,
			Total:       u.Total,
			UsedPercent: u.UsedPercent,
		})
	}
}

func (c *Collector) collectNet(s *Snapshot) {
	counters, err := net.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return
	}
	recv, sent := counters[0].BytesRecv, counters[0].BytesSent
	now := time.Now()

	if !c.lastNetTime.IsZero() {
		elapsed := now.Sub(c.lastNetTime).Seconds()
		if elapsed > 0 {
			s.NetRecvRate = float64(recv-c.lastNetRecv) / elapsed
			s.NetSentRate = float64(sent-c.lastNetSent) / elapsed
		}
	}

	c.lastNetRecv, c.lastNetSent, c.lastNetTime = recv, sent, now
}

func (c *Collector) collectProcs(s *Snapshot) {
	procs, err := process.Processes()
	if err != nil {
		return
	}

	now := time.Now()
	current := make(map[int32]procCPU, len(procs))
	list := make([]ProcInfo, 0, len(procs))

	for _, p := range procs {
		times, err := p.Times()
		if err != nil {
			continue
		}
		total := times.User + times.System

		cpuPct := 0.0
		if prev, ok := c.lastProcCPU[p.Pid]; ok {
			elapsed := now.Sub(prev.when).Seconds()
			if elapsed > 0 {
				cpuPct = (total - prev.total) / elapsed * 100.0
			}
		}
		current[p.Pid] = procCPU{total: total, when: now}

		name, _ := p.Name()
		memMB := 0.0
		if mi, err := p.MemoryInfo(); err == nil {
			memMB = float64(mi.RSS) / 1024.0 / 1024.0
		}

		list = append(list, ProcInfo{
			PID:   p.Pid,
			Name:  name,
			CPU:   cpuPct,
			MemMB: memMB,
		})
	}

	c.lastProcCPU = current

	sort.Slice(list, func(i, j int) bool {
		if list[i].CPU != list[j].CPU {
			return list[i].CPU > list[j].CPU
		}
		return list[i].MemMB > list[j].MemMB
	})

	if len(list) > MaxProcs {
		list = list[:MaxProcs]
	}
	s.Procs = list
}
