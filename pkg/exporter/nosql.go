package exporter

import (
	"encoding/json"
	"fmt"
	"os"

	"load-motion-animation/pkg/motion"
)

// ExportJSON exports motion data to a JSON document (MongoDB import format or standard document)
func ExportJSON(m *motion.MotionData, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to encode motion data: %w", err)
	}

	return nil
}

// ExportMongoLines exports motion data into newline-delimited JSON (NDJSON) for mongoimport
func ExportMongoLines(m *motion.MotionData, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create NDJSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, frame := range m.Frames {
		doc := map[string]interface{}{
			"motion_file": m.FileName,
			"format":      m.Format,
			"frame_index": frame.FrameIndex,
			"time_sec":    frame.TimeSec,
			"joints":      frame.Joints,
			"rotations":   frame.Rotations,
		}
		if err := encoder.Encode(doc); err != nil {
			return err
		}
	}

	return nil
}
