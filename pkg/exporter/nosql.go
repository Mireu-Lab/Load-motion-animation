package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"load-motion-animation/pkg/motion"
)

// ExportJSON exports the complete [motion.MotionData] structure as an indented JSON document to a file.
func ExportJSON(m *motion.MotionData, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	return ExportJSONWriter(m, file)
}

// ExportJSONWriter serializes [motion.MotionData] as indented JSON into an [io.Writer] stream.
func ExportJSONWriter(m *motion.MotionData, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to encode motion data: %w", err)
	}
	return nil
}

// ExportMongoLines exports motion data into newline-delimited JSON (NDJSON) format suitable for mongoimport.
// Each line in the output represents a single motion frame document with timestamp and joint coordinates.
func ExportMongoLines(m *motion.MotionData, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create NDJSON file: %w", err)
	}
	defer file.Close()

	return ExportMongoLinesWriter(m, file)
}

// ExportMongoLinesWriter streams newline-delimited JSON frame documents into an [io.Writer].
func ExportMongoLinesWriter(m *motion.MotionData, w io.Writer) error {
	encoder := json.NewEncoder(w)
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

// ExportStreamMongoLines streams parsed frames from a StreamParser directly to newline-delimited JSON (NDJSON) on disk.
func ExportStreamMongoLines(parser motion.StreamParser, inputPath, outputPath string) (*motion.MotionData, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create NDJSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	callback := func(frameIdx int, timeSec float64, positions map[string]motion.Vector3, rotations map[string]motion.Vector3, root *motion.Joint) error {
		doc := map[string]interface{}{
			"motion_file": inputPath,
			"frame_index": frameIdx,
			"time_sec":    timeSec,
			"joints":      positions,
			"rotations":   rotations,
		}
		return encoder.Encode(doc)
	}

	meta, err := parser.StreamFile(inputPath, callback)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
