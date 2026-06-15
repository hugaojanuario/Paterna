package alert

import "time"

type AlertRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Condition Condition `json:"condition"`
	Threshold float64   `json:"threshold"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type AlertHistory struct {
	ID            string    `json:"id"`
	RuleID        string    `json:"rule_id"`
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	Message       string    `json:"message"`
	SendAt        time.Time `json:"send_at"`
}

type Condition string

const (
	ConditionCPUHigh       Condition = "cpu_high"
	ConditionMemHigh       Condition = "mem_high"
	ConditionContainerDown Condition = "container_down"
	ConditionOOMKill       Condition = "oom_kill"
)
