package tests

import (
	"os"
	"path/filepath"
	"testing"

	"load-motion-animation/pkg/bvh"
	"load-motion-animation/pkg/c3d"
	"load-motion-animation/pkg/df"
	"load-motion-animation/pkg/exporter"
	"load-motion-animation/pkg/visualizer"
)

func TestBVHParser(t *testing.T) {
	bvhFile := filepath.Join("..", "_sample", "Sample_motion.bvh")
	if _, err := os.Stat(bvhFile); os.IsNotExist(err) {
		t.Skip("Sample BVH file not found")
	}

	parser := bvh.NewParser()
	motionData, err := parser.ParseFile(bvhFile)
	if err != nil {
		t.Fatalf("Failed to parse BVH file: %v", err)
	}

	if motionData.Format != "BVH" {
		t.Errorf("Expected format BVH, got %s", motionData.Format)
	}

	if motionData.TotalFrames <= 0 {
		t.Errorf("Expected positive frame count, got %d", motionData.TotalFrames)
	}

	dataframe := df.FromMotionData(motionData)
	if len(dataframe.Rows) == 0 {
		t.Errorf("Expected non-empty dataframe rows")
	}
}

func TestC3DParser(t *testing.T) {
	c3dFile := filepath.Join("..", "_sample", "arthuman-sample.c3d")
	if _, err := os.Stat(c3dFile); os.IsNotExist(err) {
		t.Skip("Sample C3D file not found")
	}

	parser := c3d.NewParser()
	motionData, err := parser.ParseFile(c3dFile)
	if err != nil {
		t.Fatalf("Failed to parse C3D file: %v", err)
	}

	if motionData.Format != "C3D" {
		t.Errorf("Expected format C3D, got %s", motionData.Format)
	}

	if motionData.TotalFrames <= 0 {
		t.Errorf("Expected positive frame count, got %d", motionData.TotalFrames)
	}
}

func TestExportersAndGIF(t *testing.T) {
	bvhFile := filepath.Join("..", "_sample", "Sample_motion.bvh")
	if _, err := os.Stat(bvhFile); os.IsNotExist(err) {
		t.Skip("Sample BVH file not found")
	}

	parser := bvh.NewParser()
	motionData, err := parser.ParseFile(bvhFile)
	if err != nil {
		t.Fatalf("Failed to parse BVH file: %v", err)
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	sqlPath := filepath.Join(tmpDir, "test.sql")
	gifPath := filepath.Join(tmpDir, "test.gif")

	if err := exporter.ExportCSV(motionData, csvPath); err != nil {
		t.Errorf("CSV Export failed: %v", err)
	}

	if err := exporter.ExportSQLScript(motionData, sqlPath, exporter.SQLite, "test_table"); err != nil {
		t.Errorf("SQL Export failed: %v", err)
	}

	opts := visualizer.DefaultGIFOptions()
	opts.MaxFrames = 10
	if err := visualizer.ExportGIF(motionData, gifPath, opts); err != nil {
		t.Errorf("GIF Export failed: %v", err)
	}
}
