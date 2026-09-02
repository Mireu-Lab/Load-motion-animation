package motion_test

import (
	"fmt"

	"load-motion-animation/pkg/motion"
)

func ExampleVector3() {
	v := motion.Vector3{X: 10.5, Y: 20.0, Z: -5.25}
	fmt.Println(v.String())
	// Output: (10.500, 20.000, -5.250)
}

func ExampleMotionData_Duration() {
	m := &motion.MotionData{
		Format:        "BVH",
		FrameRate:     60.0,
		TotalFrames:   120,
		FrameDuration: 1.0 / 60.0,
	}
	fmt.Printf("Duration: %.1fs\n", m.Duration())
	// Output: Duration: 2.0s
}
