package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"load-motion-animation/pkg/motion"
)

// ExportCSV exports motion frames to a CSV file
func ExportCSV(m *motion.MotionData, outputPath string) error {
	return exportDelimited(m, outputPath, ',')
}

// ExportTSV exports motion frames to a TSV file
func ExportTSV(m *motion.MotionData, outputPath string) error {
	return exportDelimited(m, outputPath, '\t')
}

func exportDelimited(m *motion.MotionData, outputPath string, comma rune) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
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
