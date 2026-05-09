package optimizer

import (
	"time"

	"github.com/affordmed/vehicle_maintenance_scheduler/internal/scheduler/domain"
)

// Optimizer defines the interface for task selection optimization.
type Optimizer interface {
	Optimize(tasks []domain.Task, availableHours int) domain.OptimizationResult
}

// KnapsackOptimizer implements the 0/1 Knapsack Dynamic Programming approach.
type KnapsackOptimizer struct{}

// NewKnapsackOptimizer creates a new Knapsack DP optimizer.
func NewKnapsackOptimizer() *KnapsackOptimizer {
	return &KnapsackOptimizer{}
}

// Optimize maximizes the total impact score within the given mechanic hour constraints.
func (o *KnapsackOptimizer) Optimize(tasks []domain.Task, availableHours int) domain.OptimizationResult {
	n := len(tasks)
	if n == 0 || availableHours <= 0 {
		return domain.OptimizationResult{SelectedTasks: []domain.Task{}, TotalImpact: 0, HoursUsed: 0}
	}

	// dp[i][w] stores the max impact score using the first i tasks and w hours
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, availableHours+1)
	}

	for i := 1; i <= n; i++ {
		task := tasks[i-1]
		hours := int(task.Duration / time.Hour) // assuming Duration is in hours

		for w := 1; w <= availableHours; w++ {
			if hours <= w {
				// We can include this task
				includeImpact := task.ImpactScore + dp[i-1][w-hours]
				excludeImpact := dp[i-1][w]

				if includeImpact > excludeImpact {
					dp[i][w] = includeImpact
				} else {
					dp[i][w] = excludeImpact
				}
			} else {
				// Cannot include, task requires more hours than w
				dp[i][w] = dp[i-1][w]
			}
		}
	}

	// Backtrack to find the selected tasks
	var selected []domain.Task
	w := availableHours
	hoursUsed := 0

	for i := n; i > 0 && w > 0; i-- {
		if dp[i][w] != dp[i-1][w] {
			// Item i was included
			task := tasks[i-1]
			selected = append(selected, task)
			h := int(task.Duration / time.Hour)
			w -= h
			hoursUsed += h
		}
	}

	// Reversing to maintain some chronological or ID order (optional)
	for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
		selected[i], selected[j] = selected[j], selected[i]
	}

	return domain.OptimizationResult{
		SelectedTasks: selected,
		TotalImpact:   dp[n][availableHours],
		HoursUsed:     hoursUsed,
	}
}
