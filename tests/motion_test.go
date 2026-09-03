package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/mireu-lab/Load-motion-animation/pkg/bvh"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/c3d"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/df"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/exporter"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/motion"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/trc"
	"gitlab.com/mireu-lab/Load-motion-animation/pkg/visualizer"
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

func TestTRCParser(t *testing.T) {
	sampleTRC := "PathFileType\t4\t(X/Y/Z)\ttest.trc\n" +
		"DataRate\tCameraRate\tNumFrames\tNumMarkers\tUnits\tOrigDataRate\tOrigDataStartFrame\tOrigNumFrames\n" +
		"60.0\t60.0\t2\t2\tmm\t60.0\t1\t2\n" +
		"Frame#\tTime\tMarker1\t\t\tMarker2\t\t\n" +
		"\t\tX1\tY1\tZ1\tX2\tY2\tZ2\n" +
		"1\t0.016667\t1.0\t2.0\t3.0\t4.0\t5.0\t6.0\n" +
		"2\t0.033333\t1.1\t2.1\t3.1\t4.1\t5.1\t6.1\n"

	parser := trc.NewParser()
	motionData, err := parser.ParseReader(strings.NewReader(sampleTRC), "test.trc")
	if err != nil {
		t.Fatalf("Failed to parse TRC: %v", err)
	}

	if motionData.Format != "TRC" {
		t.Errorf("Expected format TRC, got %s", motionData.Format)
	}
	if motionData.TotalFrames != 2 {
		t.Errorf("Expected 2 frames, got %d", motionData.TotalFrames)
	}
	if len(motionData.JointList) != 2 {
		t.Errorf("Expected 2 markers, got %d", len(motionData.JointList))
	}
}

func TestStreamParsers(t *testing.T) {
	bvhFile := filepath.Join("..", "_sample", "Sample_motion.bvh")
	if _, err := os.Stat(bvhFile); os.IsNotExist(err) {
		t.Skip("Sample BVH file not found")
	}

	bvhParser := bvh.NewParser()
	var bvhStreamFrames int
	_, err := bvhParser.StreamFile(bvhFile, func(frameIdx int, timeSec float64, positions map[string]motion.Vector3, rotations map[string]motion.Vector3, root *motion.Joint) error {
		bvhStreamFrames++
		return nil
	})
	if err != nil {
		t.Fatalf("BVH streaming failed: %v", err)
	}
	if bvhStreamFrames != 3755 {
		t.Errorf("Expected 3755 streamed BVH frames, got %d", bvhStreamFrames)
	}

	c3dFile := filepath.Join("..", "_sample", "arthuman-sample.c3d")
	if _, err := os.Stat(c3dFile); err == nil {
		c3dParser := c3d.NewParser()
		var c3dStreamFrames int
		_, err := c3dParser.StreamFile(c3dFile, func(frameIdx int, timeSec float64, positions map[string]motion.Vector3, rotations map[string]motion.Vector3, root *motion.Joint) error {
			c3dStreamFrames++
			return nil
		})
		if err != nil {
			t.Fatalf("C3D streaming failed: %v", err)
		}
		if c3dStreamFrames <= 0 {
			t.Errorf("Expected positive streamed C3D frames, got %d", c3dStreamFrames)
		}
	}
}

func TestStreamingExporters(t *testing.T) {
	bvhFile := filepath.Join("..", "_sample", "Sample_motion.bvh")
	if _, err := os.Stat(bvhFile); os.IsNotExist(err) {
		t.Skip("Sample BVH file not found")
	}

	parser := bvh.NewParser()
	tmpDir := t.TempDir()

	csvOut := filepath.Join(tmpDir, "stream.csv")
	tsvOut := filepath.Join(tmpDir, "stream.tsv")
	sqlOut := filepath.Join(tmpDir, "stream.sql")
	ndjsonOut := filepath.Join(tmpDir, "stream.ndjson")

	if _, err := exporter.ExportStreamCSV(parser, bvhFile, csvOut); err != nil {
		t.Errorf("ExportStreamCSV failed: %v", err)
	}
	if _, err := exporter.ExportStreamTSV(parser, bvhFile, tsvOut); err != nil {
		t.Errorf("ExportStreamTSV failed: %v", err)
	}
	if _, err := exporter.ExportStreamSQLScript(parser, bvhFile, sqlOut, exporter.PostgreSQL, "stream_test"); err != nil {
		t.Errorf("ExportStreamSQLScript failed: %v", err)
	}
	if _, err := exporter.ExportStreamMongoLines(parser, bvhFile, ndjsonOut); err != nil {
		t.Errorf("ExportStreamMongoLines failed: %v", err)
	}

	for _, path := range []string{csvOut, tsvOut, sqlOut, ndjsonOut} {
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Errorf("Exported file %s is missing or empty", path)
		}
	}
}
