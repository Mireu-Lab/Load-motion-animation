package bvh

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"load-motion-animation/pkg/motion"
)

// Parser handles parsing of BioVision Hierarchy (.bvh) files
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

type channelTarget struct {
	JointName   string
	ChannelType string // Xposition, Yposition, Zposition, Zrotation, Xrotation, Yrotation, etc.
}

func (p *Parser) ParseFile(filePath string) (*motion.MotionData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open BVH file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Buffer capacity for long motion data lines
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var root *motion.Joint
	var jointStack []*motion.Joint
	var channelMap []channelTarget
	var jointList []string

	inHierarchy := false
	inMotion := false

	motionData := &motion.MotionData{
		Format:   "BVH",
		FileName: filePath,
		MetaData: make(map[string]string),
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		token := fields[0]

		if token == "HIERARCHY" {
			inHierarchy = true
			continue
		}

		if token == "MOTION" {
			inHierarchy = false
			inMotion = true
			continue
		}

		if inHierarchy {
			switch token {
			case "ROOT", "JOINT":
				name := fields[1]
				joint := &motion.Joint{
					Name:     name,
					Channels: []string{},
					Children: []*motion.Joint{},
				}
				if len(jointStack) > 0 {
					parent := jointStack[len(jointStack)-1]
					joint.ParentName = parent.Name
					parent.Children = append(parent.Children, joint)
				} else {
					root = joint
				}
				jointStack = append(jointStack, joint)
				jointList = append(jointList, name)

			case "End": // End Site
				joint := &motion.Joint{
					Name:     fmt.Sprintf("%s_End", jointStack[len(jointStack)-1].Name),
					Channels: []string{},
					Children: []*motion.Joint{},
				}
				parent := jointStack[len(jointStack)-1]
				joint.ParentName = parent.Name
				parent.Children = append(parent.Children, joint)
				jointStack = append(jointStack, joint)

			case "{":
				// Bracket open

			case "}":
				// Bracket close
				if len(jointStack) > 0 {
					jointStack = jointStack[:len(jointStack)-1]
				}

			case "OFFSET":
				if len(fields) >= 4 && len(jointStack) > 0 {
					x, _ := strconv.ParseFloat(fields[1], 64)
					y, _ := strconv.ParseFloat(fields[2], 64)
					z, _ := strconv.ParseFloat(fields[3], 64)
					curr := jointStack[len(jointStack)-1]
					if strings.HasSuffix(curr.Name, "_End") {
						curr.EndSite = &motion.Vector3{X: x, Y: y, Z: z}
					} else {
						curr.Offset = motion.Vector3{X: x, Y: y, Z: z}
					}
				}

			case "CHANNELS":
				if len(fields) >= 2 && len(jointStack) > 0 {
					numChan, _ := strconv.Atoi(fields[1])
					curr := jointStack[len(jointStack)-1]
					for i := 0; i < numChan && (2+i) < len(fields); i++ {
						ch := fields[2+i]
						curr.Channels = append(curr.Channels, ch)
						channelMap = append(channelMap, channelTarget{
							JointName:   curr.Name,
							ChannelType: ch,
						})
					}
				}
			}
		} else if inMotion {
			if strings.HasPrefix(line, "Frames:") {
				if len(fields) >= 2 {
					totalFrames, _ := strconv.Atoi(fields[1])
					motionData.TotalFrames = totalFrames
				}
			} else if strings.HasPrefix(line, "Frame Time:") {
				if len(fields) >= 3 {
					frameDur, _ := strconv.ParseFloat(fields[2], 64)
					motionData.FrameDuration = frameDur
					if frameDur > 0 {
						motionData.FrameRate = 1.0 / frameDur
					}
				}
			} else {
				// Parse frame line
				if len(fields) < len(channelMap) {
					continue
				}

				frameIdx := len(motionData.Frames)
				rawVals := make([]float64, len(fields))
				jointValues := make(map[string]map[string]float64)

				for i, field := range fields {
					val, _ := strconv.ParseFloat(field, 64)
					rawVals[i] = val
					if i < len(channelMap) {
						target := channelMap[i]
						if _, exists := jointValues[target.JointName]; !exists {
							jointValues[target.JointName] = make(map[string]float64)
						}
						jointValues[target.JointName][target.ChannelType] = val
					}
				}

				// Compute forward kinematics positions for skeleton joints
				jointPositions, jointRotations := computeForwardKinematics(root, jointValues)

				frame := motion.FrameData{
					FrameIndex: frameIdx,
					TimeSec:    float64(frameIdx) * motionData.FrameDuration,
					Joints:     jointPositions,
					Rotations:  jointRotations,
					RawValues:  rawVals,
				}

				motionData.Frames = append(motionData.Frames, frame)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading BVH file: %w", err)
	}

	motionData.RootJoint = root
	motionData.JointList = jointList

	return motionData, nil
}

// computeForwardKinematics calculates 3D world space positions for all joints in a frame
func computeForwardKinematics(root *motion.Joint, values map[string]map[string]float64) (map[string]motion.Vector3, map[string]motion.Vector3) {
	positions := make(map[string]motion.Vector3)
	rotations := make(map[string]motion.Vector3)

	if root == nil {
		return positions, rotations
	}

	// 4x4 Identity transform matrix
	identity := matrix4x4Identity()
	traverseKinematics(root, identity, motion.Vector3{X: 0, Y: 0, Z: 0}, values, positions, rotations)

	return positions, rotations
}

type matrix4 [4][4]float64

func matrix4x4Identity() matrix4 {
	return matrix4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func multiplyMatrix(a, b matrix4) matrix4 {
	var res matrix4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				res[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return res
}

func translationMatrix(tx, ty, tz float64) matrix4 {
	m := matrix4x4Identity()
	m[0][3] = tx
	m[1][3] = ty
	m[2][3] = tz
	return m
}

func rotationMatrixX(deg float64) matrix4 {
	rad := deg * math.Pi / 180.0
	c, s := math.Cos(rad), math.Sin(rad)
	m := matrix4x4Identity()
	m[1][1] = c
	m[1][2] = -s
	m[2][1] = s
	m[2][2] = c
	return m
}

func rotationMatrixY(deg float64) matrix4 {
	rad := deg * math.Pi / 180.0
	c, s := math.Cos(rad), math.Sin(rad)
	m := matrix4x4Identity()
	m[0][0] = c
	m[0][2] = s
	m[2][0] = -s
	m[2][2] = c
	return m
}

func rotationMatrixZ(deg float64) matrix4 {
	rad := deg * math.Pi / 180.0
	c, s := math.Cos(rad), math.Sin(rad)
	m := matrix4x4Identity()
	m[0][0] = c
	m[0][1] = -s
	m[1][0] = s
	m[1][1] = c
	return m
}

func traverseKinematics(
	joint *motion.Joint,
	parentTransform matrix4,
	parentPos motion.Vector3,
	values map[string]map[string]float64,
	positions map[string]motion.Vector3,
	rotations map[string]motion.Vector3,
) {
	if joint == nil {
		return
	}

	jVals := values[joint.Name]

	// Translation
	tx := joint.Offset.X
	ty := joint.Offset.Y
	tz := joint.Offset.Z

	if v, ok := jVals["Xposition"]; ok {
		tx += v
	}
	if v, ok := jVals["Yposition"]; ok {
		ty += v
	}
	if v, ok := jVals["Zposition"]; ok {
		tz += v
	}

	// Translation matrix
	T := translationMatrix(tx, ty, tz)

	// Rotations
	rx := jVals["Xrotation"]
	ry := jVals["Yrotation"]
	rz := jVals["Zrotation"]

	rotations[joint.Name] = motion.Vector3{X: rx, Y: ry, Z: rz}

	// Calculate rotation matrix according to channels order
	R := matrix4x4Identity()
	for _, ch := range joint.Channels {
		switch ch {
		case "Xrotation":
			R = multiplyMatrix(R, rotationMatrixX(rx))
		case "Yrotation":
			R = multiplyMatrix(R, rotationMatrixY(ry))
		case "Zrotation":
			R = multiplyMatrix(R, rotationMatrixZ(rz))
		}
	}

	// Local transform matrix
	localTransform := multiplyMatrix(T, R)
	// Global transform matrix
	globalTransform := multiplyMatrix(parentTransform, localTransform)

	// World position from global transform
	worldPos := motion.Vector3{
		X: globalTransform[0][3],
		Y: globalTransform[1][3],
		Z: globalTransform[2][3],
	}

	positions[joint.Name] = worldPos

	for _, child := range joint.Children {
		traverseKinematics(child, globalTransform, worldPos, values, positions, rotations)
	}
}
