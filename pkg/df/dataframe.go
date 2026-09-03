// Package df provides an in-memory tabular DataFrame data structure designed for
// flattening, analyzing, and preparing 3D motion capture frames for statistical
// exploration and machine learning training pipelines.
package df

import (
	"fmt"

	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/motion"
)

// Row represents a single flattened tabular record representing a joint's 3D state in a frame.
type Row map[string]interface{}

// DataFrame represents an in-memory tabular dataset with named columns and row records.
type DataFrame struct {
	// Columns is the ordered list of column field names.
	Columns []string `json:"columns"`
	// Rows is the collection of flattened row records.
	Rows []Row `json:"rows"`
}

// FromMotionData flattens a unified [motion.MotionData] structure into a tabular [DataFrame].
// Each row corresponds to a specific joint state in a single frame, containing:
// "frame", "time_sec", "joint_name", "pos_x", "pos_y", "pos_z", "rot_x", "rot_y", and "rot_z".
func FromMotionData(m *motion.MotionData) *DataFrame {
	columns := []string{"frame", "time_sec", "joint_name", "pos_x", "pos_y", "pos_z", "rot_x", "rot_y", "rot_z"}
	rows := make([]Row, 0, len(m.Frames)*len(m.JointList))

	for _, frame := range m.Frames {
		for jointName, pos := range frame.Joints {
			rot := frame.Rotations[jointName]
			row := Row{
				"frame":      frame.FrameIndex,
				"time_sec":   frame.TimeSec,
				"joint_name": jointName,
				"pos_x":      pos.X,
				"pos_y":      pos.Y,
				"pos_z":      pos.Z,
				"rot_x":      rot.X,
				"rot_y":      rot.Y,
				"rot_z":      rot.Z,
			}
			rows = append(rows, row)
		}
	}

	return &DataFrame{
		Columns: columns,
		Rows:    rows,
	}
}

// Len returns the total number of rows in the DataFrame.
func (df *DataFrame) Len() int {
	return len(df.Rows)
}

// Head returns the first n rows of the DataFrame.
func (df *DataFrame) Head(n int) []Row {
	if n > len(df.Rows) {
		n = len(df.Rows)
	}
	if n < 0 {
		n = 0
	}
	return df.Rows[:n]
}

// Summary returns a human-readable dimension description of the DataFrame (e.g., "DataFrame: 500 rows x 9 columns").
func (df *DataFrame) Summary() string {
	return fmt.Sprintf("DataFrame: %d rows x %d columns", len(df.Rows), len(df.Columns))
}
