package df_test

import (
	"fmt"

	"load-motion-animation/pkg/df"
	"load-motion-animation/pkg/motion"
)

func ExampleFromMotionData() {
	m := &motion.MotionData{
		Format:      "BVH",
		JointList:   []string{"Hips"},
		TotalFrames: 1,
		Frames: []motion.FrameData{
			{
				FrameIndex: 0,
				TimeSec:    0.0,
				Joints: map[string]motion.Vector3{
					"Hips": {X: 0.0, Y: 1.0, Z: 0.0},
				},
				Rotations: map[string]motion.Vector3{
					"Hips": {X: 0.0, Y: 0.0, Z: 0.0},
				},
			},
		},
	}

	dataframe := df.FromMotionData(m)
	fmt.Println(dataframe.Summary())
	// Output: DataFrame: 1 rows x 9 columns
}
