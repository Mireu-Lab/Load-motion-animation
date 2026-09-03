package exporter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/motion"
)

// RDBMSType specifies the target relational database SQL dialect.
type RDBMSType string

const (
	// PostgreSQL dialect generates DDL with BIGSERIAL primary keys and DOUBLE PRECISION coordinates.
	PostgreSQL RDBMSType = "postgresql"
	// MySQL dialect generates DDL with BIGINT AUTO_INCREMENT and DOUBLE types.
	MySQL RDBMSType = "mysql"
	// SQLite dialect generates DDL with INTEGER PRIMARY KEY AUTOINCREMENT and REAL coordinate types.
	SQLite RDBMSType = "sqlite"
)

// ExportSQLScript generates a complete SQL transaction script containing schema DDL and
// batch INSERT statements for all motion frames, writing the output to a file.
// If tableName is empty, "motion_frames" is used by default.
func ExportSQLScript(m *motion.MotionData, outputPath string, dbType RDBMSType, tableName string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create SQL script file: %w", err)
	}
	defer file.Close()

	return ExportSQLScriptWriter(m, file, dbType, tableName)
}

// ExportSQLScriptWriter writes DDL and multi-row INSERT SQL statements into an [io.Writer] stream.
func ExportSQLScriptWriter(m *motion.MotionData, w io.Writer, dbType RDBMSType, tableName string) error {
	if tableName == "" {
		tableName = "motion_frames"
	}

	// Write Table Schema Creation
	var schema string
	switch dbType {
	case PostgreSQL:
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id BIGSERIAL PRIMARY KEY,
    motion_file VARCHAR(255),
    frame_index INT,
    time_sec DOUBLE PRECISION,
    joint_name VARCHAR(100),
    pos_x DOUBLE PRECISION,
    pos_y DOUBLE PRECISION,
    pos_z DOUBLE PRECISION,
    rot_x DOUBLE PRECISION,
    rot_y DOUBLE PRECISION,
    rot_z DOUBLE PRECISION
);
`, tableName)
	case MySQL:
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    motion_file VARCHAR(255),
    frame_index INT,
    time_sec DOUBLE,
    joint_name VARCHAR(100),
    pos_x DOUBLE,
    pos_y DOUBLE,
    pos_z DOUBLE,
    rot_x DOUBLE,
    rot_y DOUBLE,
    rot_z DOUBLE
);
`, tableName)
	default: // SQLite
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    motion_file TEXT,
    frame_index INTEGER,
    time_sec REAL,
    joint_name TEXT,
    pos_x REAL,
    pos_y REAL,
    pos_z REAL,
    rot_x REAL,
    rot_y REAL,
    rot_z REAL
);
`, tableName)
	}

	if _, err := io.WriteString(w, schema+"\nBEGIN TRANSACTION;\n"); err != nil {
		return err
	}

	batchSize := 500
	valueTuples := make([]string, 0, batchSize)

	insertHeader := fmt.Sprintf("INSERT INTO %s (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n", tableName)

	writeBatch := func() error {
		if len(valueTuples) == 0 {
			return nil
		}
		stmt := insertHeader + strings.Join(valueTuples, ",\n") + ";\n"
		if _, err := io.WriteString(w, stmt); err != nil {
			return err
		}
		valueTuples = valueTuples[:0]
		return nil
	}

	fileNameEsc := strings.ReplaceAll(m.FileName, "'", "''")

	for _, frame := range m.Frames {
		for jointName, pos := range frame.Joints {
			rot := frame.Rotations[jointName]
			jointEsc := strings.ReplaceAll(jointName, "'", "''")
			tuple := fmt.Sprintf("('%s', %d, %.6f, '%s', %.6f, %.6f, %.6f, %.6f, %.6f, %.6f)",
				fileNameEsc, frame.FrameIndex, frame.TimeSec, jointEsc, pos.X, pos.Y, pos.Z, rot.X, rot.Y, rot.Z)
			valueTuples = append(valueTuples, tuple)

			if len(valueTuples) >= batchSize {
				if err := writeBatch(); err != nil {
					return err
				}
			}
		}
	}

	if err := writeBatch(); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "COMMIT;\n"); err != nil {
		return err
	}

	return nil
}

// ExportStreamSQLScript streams motion frames from a StreamParser directly to a SQL file without buffering all frames in memory.
func ExportStreamSQLScript(parser motion.StreamParser, inputPath, outputPath string, dbType RDBMSType, tableName string) (*motion.MotionData, error) {
	if tableName == "" {
		tableName = "motion_frames"
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL script file: %w", err)
	}
	defer file.Close()

	var schema string
	switch dbType {
	case PostgreSQL:
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id BIGSERIAL PRIMARY KEY,
    motion_file VARCHAR(255),
    frame_index INT,
    time_sec DOUBLE PRECISION,
    joint_name VARCHAR(100),
    pos_x DOUBLE PRECISION,
    pos_y DOUBLE PRECISION,
    pos_z DOUBLE PRECISION,
    rot_x DOUBLE PRECISION,
    rot_y DOUBLE PRECISION,
    rot_z DOUBLE PRECISION
);
`, tableName)
	case MySQL:
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    motion_file VARCHAR(255),
    frame_index INT,
    time_sec DOUBLE,
    joint_name VARCHAR(100),
    pos_x DOUBLE,
    pos_y DOUBLE,
    pos_z DOUBLE,
    rot_x DOUBLE,
    rot_y DOUBLE,
    rot_z DOUBLE
);
`, tableName)
	default: // SQLite
		schema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    motion_file TEXT,
    frame_index INTEGER,
    time_sec REAL,
    joint_name TEXT,
    pos_x REAL,
    pos_y REAL,
    pos_z REAL,
    rot_x REAL,
    rot_y REAL,
    rot_z REAL
);
`, tableName)
	}

	if _, err := io.WriteString(file, schema+"\nBEGIN TRANSACTION;\n"); err != nil {
		return nil, err
	}

	batchSize := 500
	valueTuples := make([]string, 0, batchSize)
	insertHeader := fmt.Sprintf("INSERT INTO %s (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n", tableName)

	writeBatch := func() error {
		if len(valueTuples) == 0 {
			return nil
		}
		stmt := insertHeader + strings.Join(valueTuples, ",\n") + ";\n"
		if _, err := io.WriteString(file, stmt); err != nil {
			return err
		}
		valueTuples = valueTuples[:0]
		return nil
	}

	fileNameEsc := strings.ReplaceAll(inputPath, "'", "''")

	callback := func(frameIdx int, timeSec float64, positions map[string]motion.Vector3, rotations map[string]motion.Vector3, root *motion.Joint) error {
		for jointName, pos := range positions {
			rot := rotations[jointName]
			jointEsc := strings.ReplaceAll(jointName, "'", "''")
			tuple := fmt.Sprintf("('%s', %d, %.6f, '%s', %.6f, %.6f, %.6f, %.6f, %.6f, %.6f)",
				fileNameEsc, frameIdx, timeSec, jointEsc, pos.X, pos.Y, pos.Z, rot.X, rot.Y, rot.Z)
			valueTuples = append(valueTuples, tuple)

			if len(valueTuples) >= batchSize {
				if err := writeBatch(); err != nil {
					return err
				}
			}
		}
		return nil
	}

	meta, err := parser.StreamFile(inputPath, callback)
	if err != nil {
		return nil, err
	}

	if err := writeBatch(); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(file, "COMMIT;\n"); err != nil {
		return nil, err
	}

	return meta, nil
}
