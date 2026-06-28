package service

import (
	"context"
	"testing"
)

func TestOrchestratorBasic(t *testing.T) {
	ctx := context.Background()
	orch := NewOrchestrator(ctx)

	if orch.activeTab != "activity" {
		t.Errorf("Expected initial activeTab to be activity, got %s", orch.activeTab)
	}

	orch.SetActiveTab("plugins")
	if orch.activeTab != "plugins" {
		t.Errorf("Expected activeTab to be plugins, got %s", orch.activeTab)
	}

	info := orch.GetDetailedInfo("Apache")
	if info.Name != "Apache" {
		t.Errorf("Expected name Apache, got %s", info.Name)
	}
	if info.Status != "Stopped" {
		t.Errorf("Expected status Stopped, got %s", info.Status)
	}

	if orch.IsRunning("Apache") {
		t.Error("Expected Apache NOT to be running")
	}
}

func TestOrchestratorRefresh(t *testing.T) {
	orch := NewOrchestrator(context.Background())
	orch.needsRefresh = false
	orch.RequestRefresh()
	if !orch.needsRefresh {
		t.Error("Expected needsRefresh to be true after RequestRefresh")
	}
}
