package screenshare

import "testing"

func TestBuildDefaults(t *testing.T) {
	c, err := Build(Options{Enabled: true, Preset: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Width != 1920 || c.Height != 1080 || c.FPS != 30 || c.BitrateMbps != 8 {
		t.Fatalf("unexpected preset values: %+v", c)
	}
	if c.H264Level != "4.1" {
		t.Fatalf("expected level 4.1, got %s", c.H264Level)
	}
}

func TestBuildOverrides(t *testing.T) {
	c, err := Build(Options{
		Enabled:     true,
		Preset:      "medium",
		Width:       1280,
		Height:      720,
		FPS:         60,
		BitrateMbps: 12,
		Source:      "desktop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.FPS != 60 || c.BitrateMbps != 12 {
		t.Fatalf("overrides not applied: %+v", c)
	}
	if c.H264Level != "4.2" {
		t.Fatalf("expected level 4.2 for 60 fps, got %s", c.H264Level)
	}
}

func TestBuildRejectsInvalid(t *testing.T) {
	tests := []Options{
		{Enabled: true, Preset: "nope"},
		{Enabled: true, Width: 1281},
		{Enabled: true, FPS: 61},
		{Enabled: true, BitrateMbps: -1, Preset: "medium"},
		{Enabled: true, Source: "window"},
	}
	for i, o := range tests {
		if _, err := Build(o); err == nil {
			t.Fatalf("case %d should have failed", i)
		}
	}
}

func TestBuildDisabled(t *testing.T) {
	c, err := Build(Options{})
	if err != nil || c.Enabled {
		t.Fatalf("disabled config should be empty: %+v, %v", c, err)
	}
}
