// Package exporter provides serialization and data storage integration for motion capture datasets.
//
// It supports exporting parsed 3D motion frames into various formats including
// CSV/TSV flat files, JSON/NDJSON documents for NoSQL (such as MongoDB import),
// and batch SQL INSERT script generation for relational databases (PostgreSQL, MySQL, SQLite).
package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/motion"
)

// ExportCSV exports motion frames to a comma-separated values (CSV) file on disk.
func ExportCSV(m *motion.MotionData, outputPath string) error {
	return exportDelimited(m, outputPath, ',')
}

// ExportTSV exports motion frames to a tab-separated values (TSV) file on disk.
func ExportTSV(m *motion.MotionData, outputPath string) error {
	return exportDelimited(m, outputPath, '\t')
}

// ExportCSVWriter writes CSV formatted motion frame rows directly into an [io.Writer] stream.
func ExportCSVWriter(m *motion.MotionData, w io.Writer) error {
	return writeDelimited(m, w, ',')
}

// ExportTSVWriter writes TSV formatted motion frame rows directly into an [io.Writer] stream.
func ExportTSVWriter(m *motion.MotionData, w io.Writer) error {
	return writeDelimited(m, w, '\t')
}

func exportDelimited(m *motion.MotionData, outputPath string, comma rune) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	return writeDelimited(m, file, comma)
}

func writeDelimited(m *motion.MotionData, w io.Writer, comma rune) error {
	writer := csv.NewWriter(w)
	writer.Comma = comma
	defer writer.Flush()

	// Header row
	header := []string{"frame", "time_sec", "joint_name", "pos_x", "pos_y", "pos_z", "rot_x", "rot_y", "rot_z"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, frame := range m.Frames {
		for jointName, pos := range frame.Joints {
			rot := frame.Rotations[jointName]
			row := []string{
				strconv.Itoa(frame.FrameIndex),
				strconv.FormatFloat(frame.TimeSec, 'f', 6, 64),
				jointName,
				strconv.FormatFloat(pos.X, 'f', 6, 64),
				strconv.FormatFloat(pos.Y, 'f', 6, 64),
				strconv.FormatFloat(pos.Z, 'f', 6, 64),
				strconv.FormatFloat(rot.X, 'f', 6, 64),
				strconv.FormatFloat(rot.Y, 'f', 6, 64),
				strconv.FormatFloat(rot.Z, 'f', 6, 64),
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExportStreamCSV streams motion frames from parser directly to a CSV file without buffering all frames in memory.
func ExportStreamCSV(parser motion.StreamParser, inputPath, outputPath string) (*motion.MotionData, error) {
	return exportStreamDelimited(parser, inputPath, outputPath, ',')
}

// ExportStreamTSV streams motion frames from parser directly to a TSV file without buffering all frames in memory.
func ExportStreamTSV(parser motion.StreamParser, inputPath, outputPath string) (*motion.MotionData, error) {
	return exportStreamDelimited(parser, inputPath, outputPath, '\t')
}

func exportStreamDelimited(parser motion.StreamParser, inputPath, outputPath string, comma rune) (*motion.MotionData, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = comma

	header := []string{"frame", "time_sec", "joint_name", "pos_x", "pos_y", "pos_z", "rot_x", "rot_y", "rot_z"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	callback := func(frameIdx int, timeSec float64, positions map[string]motion.Vector3, rotations map[string]motion.Vector3, root *motion.Joint) error {
		for jointName, pos := range positions {
			rot := rotations[jointName]
			row := []string{
				strconv.Itoa(frameIdx),
				strconv.FormatFloat(timeSec, 'f', 6, 64),
				jointName,
				strconv.FormatFloat(pos.X, 'f', 6, 64),
				strconv.FormatFloat(pos.Y, 'f', 6, 64),
				strconv.FormatFloat(pos.Z, 'f', 6, 64),
				strconv.FormatFloat(rot.X, 'f', 6, 64),
				strconv.FormatFloat(rot.Y, 'f', 6, 64),
				strconv.FormatFloat(rot.Z, 'f', 6, 64),
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		return nil
	}

	meta, err := parser.StreamFile(inputPath, callback)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return meta, nil
}
