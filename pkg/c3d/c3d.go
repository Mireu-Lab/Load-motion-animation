package c3d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"

	"load-motion-animation/pkg/motion"
)

// Parser parses binary C3D motion capture files
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// C3DHeader represents the parsed header of a C3D binary file
type C3DHeader struct {
	ParamStartBlock uint8
	Magic           uint8
	NumPoints       uint16
	NumAnalog       uint16
	StartFrame      uint16
	EndFrame        uint16
	MaxGap          uint16
	ScaleFactor     float32
	DataStartBlock  uint16
	AnalogRate      uint16
	FrameRate       float32
}

func (p *Parser) ParseFile(filePath string) (*motion.MotionData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read C3D file: %w", err)
	}

	if len(data) < 512 {
		return nil, fmt.Errorf("file too small for C3D header")
	}

	header := C3DHeader{
		ParamStartBlock: data[0],
		Magic:           data[1],
		NumPoints:       binary.LittleEndian.Uint16(data[2:4]),
		NumAnalog:       binary.LittleEndian.Uint16(data[4:6]),
		StartFrame:      binary.LittleEndian.Uint16(data[6:8]),
		EndFrame:        binary.LittleEndian.Uint16(data[8:10]),
		MaxGap:          binary.LittleEndian.Uint16(data[10:12]),
		ScaleFactor:     math.Float32frombits(binary.LittleEndian.Uint32(data[12:16])),
		DataStartBlock:  binary.LittleEndian.Uint16(data[16:18]),
		AnalogRate:      binary.LittleEndian.Uint16(data[18:20]),
		FrameRate:       math.Float32frombits(binary.LittleEndian.Uint32(data[20:24])),
	}

	if header.Magic != 80 {
		return nil, fmt.Errorf("invalid C3D magic byte: expected 80, got %d", header.Magic)
	}

	totalFrames := int(header.EndFrame - header.StartFrame + 1)
	if totalFrames <= 0 {
		totalFrames = 0
	}

	// Read Point Labels from parameter section
	pointNames := parseC3DLabels(data, header.ParamStartBlock, header.NumPoints)

	motionData := &motion.MotionData{
		Format:      "C3D",
		FileName:    filePath,
		JointList:   pointNames,
		FrameRate:   float64(header.FrameRate),
		TotalFrames: totalFrames,
		Frames:      make([]motion.FrameData, 0, totalFrames),
		MetaData:    make(map[string]string),
	}

	if header.FrameRate > 0 {
		motionData.FrameDuration = 1.0 / float64(header.FrameRate)
	}

	// Parse 3D data frames
	dataOffset := int(header.DataStartBlock-1) * 512
	if dataOffset < 512 || dataOffset >= len(data) {
		dataOffset = 512 * 2 // Fallback
	}

	scale := header.ScaleFactor
	isFloat := scale < 0

	bytesPerPoint := 8 // 4 * int16
	if isFloat {
		bytesPerPoint = 16 // 4 * float32
	}

	frameSize := int(header.NumPoints)*bytesPerPoint + int(header.NumAnalog)*int(header.AnalogRate)*2
	if isFloat {
		frameSize = int(header.NumPoints)*bytesPerPoint + int(header.NumAnalog)*int(header.AnalogRate)*4
	}

	if frameSize <= 0 {
		return motionData, nil
	}

	currPos := dataOffset
	for f := 0; f < totalFrames && (currPos+frameSize) <= len(data); f++ {
		frameJoints := make(map[string]motion.Vector3)
		rawVals := make([]float64, 0, header.NumPoints*3)

		for i := 0; i < int(header.NumPoints); i++ {
			ptOffset := currPos + i*bytesPerPoint
			var x, y, z float64

			if isFloat {
				x = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset : ptOffset+4])))
				y = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset+4 : ptOffset+8])))
				z = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset+8 : ptOffset+12])))
			} else {
				ix := int16(binary.LittleEndian.Uint16(data[ptOffset : ptOffset+2]))
				iy := int16(binary.LittleEndian.Uint16(data[ptOffset+2 : ptOffset+4]))
				iz := int16(binary.LittleEndian.Uint16(data[ptOffset+4 : ptOffset+6]))
				sc := math.Abs(float64(scale))
				x = float64(ix) * sc
				y = float64(iy) * sc
				z = float64(iz) * sc
			}

			ptName := pointNames[i]
			frameJoints[ptName] = motion.Vector3{X: x, Y: y, Z: z}
			rawVals = append(rawVals, x, y, z)
		}

		frame := motion.FrameData{
			FrameIndex: f,
			TimeSec:    float64(f) * motionData.FrameDuration,
			Joints:     frameJoints,
			Rotations:  make(map[string]motion.Vector3),
			RawValues:  rawVals,
		}

		motionData.Frames = append(motionData.Frames, frame)
		currPos += frameSize
	}

	return motionData, nil
}

func parseC3DLabels(data []byte, paramStart uint8, numPoints uint16) []string {
	labels := make([]string, numPoints)
	for i := range labels {
		labels[i] = fmt.Sprintf("Point_%d", i+1)
	}

	paramOffset := int(paramStart-1) * 512
	if paramOffset >= len(data) {
		return labels
	}

	// Scan parameter section for LABELS string
	idx := bytes.Index(data[paramOffset:], []byte("LABELS"))
	if idx == -1 {
		return labels
	}

	labelsPos := paramOffset + idx + 6
	if labelsPos+6 >= len(data) {
		return labels
	}

	// Skip param description offset (2 bytes)
	// Read dimension info: num_dims, dim1 (str len), dim2 (num items)
	// In C3D specs: [offset_to_next (2), num_dims (1), dim1 (1), dim2 (2 or 1)]
	pos := labelsPos
	// Find first byte with string length (e.g. 8)
	for pos < labelsPos+20 && pos < len(data) {
		if data[pos] > 0 && data[pos] <= 64 { // plausible string length
			strLen := int(data[pos])
			if pos+2 < len(data) {
				count := int(binary.LittleEndian.Uint16(data[pos+1 : pos+3]))
				if count == int(numPoints) || (count > 0 && count <= 500) {
					startStr := pos + 3
					for i := 0; i < count && i < int(numPoints); i++ {
						sOffset := startStr + i*strLen
						if sOffset+strLen <= len(data) {
							lbl := strings.TrimSpace(string(data[sOffset : sOffset+strLen]))
							lbl = strings.Map(func(r rune) rune {
								if r >= 32 && r <= 126 {
									return r
								}
								return -1
							}, lbl)
							if lbl != "" {
								labels[i] = lbl
							}
						}
					}
					return labels
				}
			}
		}
		pos++
	}

	return labels
}
