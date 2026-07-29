package visualizer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"

	"load-motion-animation/pkg/motion"
)

// GIFOptions specifies options for generating GIF animations
type GIFOptions struct {
	Width       int
	Height      int
	FPS         int
	Scale       float64
	Padding     int
	MaxFrames   int // 0 for all frames
	LineColor   color.RGBA
	PointColor  color.RGBA
	BGColor     color.RGBA
}

// DefaultGIFOptions returns sensible default options
func DefaultGIFOptions() GIFOptions {
	return GIFOptions{
		Width:      400,
		Height:     400,
		FPS:        20,
		Scale:      1.0,
		Padding:    40,
		MaxFrames:  120, // max 120 frames for fast render & compact file size
		LineColor:  color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}, // Modern Blue
		PointColor: color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff}, // Vibrant Red
		BGColor:    color.RGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff}, // Dark Slate
	}
}

// ExportGIF renders motion skeleton trajectories into an animated GIF
func ExportGIF(m *motion.MotionData, outputPath string, opts GIFOptions) error {
	if len(m.Frames) == 0 {
		return fmt.Errorf("no frames to export to GIF")
	}

	if opts.Width <= 0 {
		opts.Width = 400
	}
	if opts.Height <= 0 {
		opts.Height = 400
	}

	// Determine bounding box across all frames for normalized 2D camera projection
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	numFramesToProcess := len(m.Frames)
	if opts.MaxFrames > 0 && opts.MaxFrames < numFramesToProcess {
		numFramesToProcess = opts.MaxFrames
	}

	step := 1
	if len(m.Frames) > 300 && opts.MaxFrames > 0 {
		step = len(m.Frames) / opts.MaxFrames
		if step < 1 {
			step = 1
		}
	}

	for fIdx := 0; fIdx < len(m.Frames); fIdx += step {
		frame := m.Frames[fIdx]
		for _, pos := range frame.Joints {
			if pos.X < minX {
				minX = pos.X
			}
			if pos.X > maxX {
				maxX = pos.X
			}
			if pos.Y < minY {
				minY = pos.Y
			}
			if pos.Y > maxY {
				maxY = pos.Y
			}
		}
	}

	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX == 0 {
		rangeX = 1.0
	}
	if rangeY == 0 {
		rangeY = 1.0
	}

	// Prepare Palette for GIF
	palette := []color.Color{
		opts.BGColor,
		opts.LineColor,
		opts.PointColor,
		color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // White
		color.RGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}, // Muted Gray
	}

	outGIF := &gif.GIF{}
	delay := 100 / opts.FPS
	if delay < 1 {
		delay = 1
	}

	// Build parent-child bone pairs from root joint
	bonePairs := extractBonePairs(m.RootJoint)

	for fIdx := 0; fIdx < len(m.Frames) && len(outGIF.Image) < opts.MaxFrames; fIdx += step {
		frame := m.Frames[fIdx]

		palettedImg := image.NewPaletted(image.Rect(0, 0, opts.Width, opts.Height), palette)
		// Fill Background
		draw.Draw(palettedImg, palettedImg.Bounds(), &image.Uniform{opts.BGColor}, image.Point{}, draw.Src)

		// 2D Projection helper
		project := func(v motion.Vector3) (int, int) {
			// Normalize to [0, 1]
			nx := (v.X - minX) / rangeX
			ny := (v.Y - minY) / rangeY

			// Map to Canvas pixels (invert Y for computer graphics screen coords)
			px := opts.Padding + int(nx*float64(opts.Width-2*opts.Padding))
			py := opts.Height - (opts.Padding + int(ny*float64(opts.Height-2*opts.Padding)))
			return px, py
		}

		// Draw Bones
		for _, pair := range bonePairs {
			p1, ok1 := frame.Joints[pair[0]]
			p2, ok2 := frame.Joints[pair[1]]
			if ok1 && ok2 {
				x1, y1 := project(p1)
				x2, y2 := project(p2)
				drawLine(palettedImg, x1, y1, x2, y2, opts.LineColor)
			}
		}

		// Draw Joint Nodes / Markers
		for _, pos := range frame.Joints {
			px, py := project(pos)
			drawCircle(palettedImg, px, py, 3, opts.PointColor)
		}

		outGIF.Image = append(outGIF.Image, palettedImg)
		outGIF.Delay = append(outGIF.Delay, delay)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create GIF file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, outGIF); err != nil {
		return fmt.Errorf("failed to encode GIF: %w", err)
	}

	return nil
}

func extractBonePairs(joint *motion.Joint) [][2]string {
	var pairs [][2]string
	if joint == nil {
		return pairs
	}

	var walk func(j *motion.Joint)
	walk = func(j *motion.Joint) {
		for _, child := range j.Children {
			pairs = append(pairs, [2]string{j.Name, child.Name})
			walk(child)
		}
	}
	walk(joint)

	return pairs
}

// Bresenham's line algorithm
func drawLine(img *image.Paletted, x0, y0, x1, y1 int, col color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, col)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func drawCircle(img *image.Paletted, cx, cy, r int, col color.Color) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				x, y := cx+dx, cy+dy
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, col)
				}
			}
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
