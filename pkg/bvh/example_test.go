package bvh_test

import (
	"fmt"
	"strings"

	"github.com/Mireu-Lab/Load-motion-animation/pkg/bvh"
)

func ExampleParser_ParseReader() {
	bvhContent := `HIERARCHY
ROOT Hips
{
  OFFSET 0.00 0.00 0.00
  CHANNELS 6 Xposition Yposition Zposition Zrotation Xrotation Yrotation
  End Site
  {
    OFFSET 0.00 5.00 0.00
  }
}
MOTION
Frames: 2
Frame Time: 0.033333
0.0 0.0 0.0 0.0 0.0 0.0
1.0 2.0 3.0 10.0 20.0 30.0
`
	parser := bvh.NewParser()
	data, err := parser.ParseReader(strings.NewReader(bvhContent), "sample.bvh")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Format: %s, Frames: %d\n", data.Format, data.TotalFrames)
	// Output: Format: BVH, Frames: 2
}
