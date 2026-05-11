package speedtest

import "testing"

func TestQuickPlan_NoUploadSteps(t *testing.T) {
	ups := QuickPlan.UploadSteps()
	if len(ups) != 0 {
		t.Errorf("quick plan should have no upload steps, got %d", len(ups))
	}
}

func TestQuickPlan_HasDownloadSteps(t *testing.T) {
	downs := QuickPlan.DownloadSteps()
	if len(downs) == 0 {
		t.Error("quick plan should have download steps")
	}
}

func TestDefaultPlan_HasBothDirections(t *testing.T) {
	downs := DefaultPlan.DownloadSteps()
	ups := DefaultPlan.UploadSteps()
	if len(downs) == 0 {
		t.Error("default plan should have download steps")
	}
	if len(ups) == 0 {
		t.Error("default plan should have upload steps")
	}
}

func TestThoroughPlan_HasBothDirections(t *testing.T) {
	downs := ThoroughPlan.DownloadSteps()
	ups := ThoroughPlan.UploadSteps()
	if len(downs) == 0 {
		t.Error("thorough plan should have download steps")
	}
	if len(ups) == 0 {
		t.Error("thorough plan should have upload steps")
	}
}

func TestThoroughPlan_IncludesRaw(t *testing.T) {
	if !ThoroughPlan.IncludeRaw {
		t.Error("thorough plan should include raw data")
	}
}

func TestAllPlans_PositiveBytesAndCounts(t *testing.T) {
	plans := []TestPlan{QuickPlan, DefaultPlan, ThoroughPlan}
	for _, p := range plans {
		for i, s := range p.Steps {
			if s.Bytes == 0 {
				t.Errorf("plan %s step %d has zero bytes", p.Name, i)
			}
			if s.Count == 0 {
				t.Errorf("plan %s step %d has zero count", p.Name, i)
			}
			if s.Direction != "download" && s.Direction != "upload" {
				t.Errorf("plan %s step %d has invalid direction %q", p.Name, i, s.Direction)
			}
		}
	}
}

func TestAllPlans_PositiveLatencyCount(t *testing.T) {
	plans := []TestPlan{QuickPlan, DefaultPlan, ThoroughPlan}
	for _, p := range plans {
		if p.LatencyCount <= 0 {
			t.Errorf("plan %s has non-positive latency count: %d", p.Name, p.LatencyCount)
		}
	}
}
