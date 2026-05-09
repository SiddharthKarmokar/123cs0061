package optimizer

import (
	"testing"
	"time"

	"github.com/affordmed/vehicle_maintenance_scheduler/internal/scheduler/domain"
)

func TestKnapsackOptimizer_Optimize(t *testing.T) {
	optimizer := NewKnapsackOptimizer()

	tasks := []domain.Task{
		{ID: "T1", Duration: 2 * time.Hour, ImpactScore: 50},
		{ID: "T2", Duration: 3 * time.Hour, ImpactScore: 60},
		{ID: "T3", Duration: 4 * time.Hour, ImpactScore: 80},
		{ID: "T4", Duration: 5 * time.Hour, ImpactScore: 90},
	}

	// Case 1: Max capacity 8 hours
	// Best combo: T2 (3h, 60), T4 (5h, 90) = 8h, Impact 150
	// OR T3(4h, 80) + T1(2h, 50) = 6h, Impact 130
	// OR T3(4h, 80) + T2(3h, 60) = 7h, Impact 140
	// T4(5h) + T2(3h) = 150 impact!
	res := optimizer.Optimize(tasks, 8)

	if res.TotalImpact != 150 {
		t.Errorf("Expected impact 150, got %d", res.TotalImpact)
	}
	if res.HoursUsed != 8 {
		t.Errorf("Expected hours used 8, got %d", res.HoursUsed)
	}
	if len(res.SelectedTasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(res.SelectedTasks))
	}

	// Case 2: Zero capacity
	resZero := optimizer.Optimize(tasks, 0)
	if resZero.TotalImpact != 0 || len(resZero.SelectedTasks) != 0 {
		t.Errorf("Expected empty result for zero capacity")
	}

	// Case 3: All tasks fit
	resAll := optimizer.Optimize(tasks, 20)
	if resAll.TotalImpact != 280 { // 50+60+80+90
		t.Errorf("Expected 280 impact, got %d", resAll.TotalImpact)
	}
	if len(resAll.SelectedTasks) != 4 {
		t.Errorf("Expected 4 tasks, got %d", len(resAll.SelectedTasks))
	}
}
