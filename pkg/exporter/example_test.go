package exporter_test

import (
	"bytes"
	"fmt"
	"strings"

	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/exporter"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/motion"
)

func ExampleExportCSVWriter() {
	m := &motion.MotionData{
		Frames: []motion.FrameData{
			{
				FrameIndex: 0,
				TimeSec:    0.0,
				Joints: map[string]motion.Vector3{
					"Root": {X: 1.0, Y: 2.0, Z: 3.0},
				},
				Rotations: map[string]motion.Vector3{
					"Root": {X: 0.0, Y: 0.0, Z: 0.0},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := exporter.ExportCSVWriter(m, &buf); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	fmt.Println(lines[0])
	// Output: frame,time_sec,joint_name,pos_x,pos_y,pos_z,rot_x,rot_y,rot_z
}

func ExampleExportSQLScriptWriter() {
	m := &motion.MotionData{
		FileName: "sample.bvh",
		Frames: []motion.FrameData{
			{
				FrameIndex: 0,
				TimeSec:    0.0,
				Joints: map[string]motion.Vector3{
					"Root": {X: 0.0, Y: 0.0, Z: 0.0},
				},
				Rotations: map[string]motion.Vector3{
					"Root": {X: 0.0, Y: 0.0, Z: 0.0},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := exporter.ExportSQLScriptWriter(m, &buf, exporter.SQLite, "motions"); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	hasCreateTable := strings.Contains(buf.String(), "CREATE TABLE IF NOT EXISTS motions")
	fmt.Printf("Has CREATE TABLE: %t\n", hasCreateTable)
	// Output: Has CREATE TABLE: true
}
