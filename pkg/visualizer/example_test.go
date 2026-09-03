package visualizer_test

import (
	"fmt"

	"gitlab.com/mireu-lab/Load-motion-animation/pkg/visualizer"
)

func ExampleDefaultGIFOptions() {
	opts := visualizer.DefaultGIFOptions()
	fmt.Printf("Default GIF size: %dx%d, FPS: %d\n", opts.Width, opts.Height, opts.FPS)
	// Output: Default GIF size: 400x400, FPS: 20
}
