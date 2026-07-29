package df

import (
	"fmt"

	"load-motion-animation/pkg/motion"
)

// Row represents a flat tabular row for motion frame joint data
type Row map[string]interface{}

// DataFrame represents a memory tabular data structure suitable for analysis
type DataFrame struct {
	Columns []string `json:"columns"`
	Rows    []Row    `json:"rows"`
}

// FromMotionData flattens MotionData into a DataFrame structure
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

func (df *DataFrame) Summary() string {
	return fmt.Sprintf("DataFrame: %d rows x %d columns", len(df.Rows), len(df.Columns))
}
