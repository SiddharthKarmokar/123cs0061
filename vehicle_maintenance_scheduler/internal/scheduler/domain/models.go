package domain

import "time"

// Task represents a vehicle maintenance task.
type Task struct {
	ID             string        `json:"id"`
	VehicleID      string        `json:"vehicleId"`
	Type           string        `json:"type"`
	Duration       time.Duration `json:"durationHours"` // Time required to complete the task
	ImpactScore    int           `json:"impactScore"`   // Priority/Impact of completing this task
	RequiredSkills []string      `json:"requiredSkills"`
}

// Depot represents a maintenance facility.
type Depot struct {
	ID                   string `json:"id"`
	Location             string `json:"location"`
	TotalMechanicHours   int    `json:"totalMechanicHours"`
	AvailableMechanicHrs int    `json:"availableMechanicHours"`
}

// OptimizationResult holds the outcome of the knapsack scheduling.
type OptimizationResult struct {
	SelectedTasks []Task `json:"selectedTasks"`
	TotalImpact   int    `json:"totalImpact"`
	HoursUsed     int    `json:"hoursUsed"`
}
