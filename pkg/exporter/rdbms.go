package exporter

import (
	"fmt"
	"os"
	"strings"

	"load-motion-animation/pkg/motion"
)

// RDBMSType specifies target dialect
type RDBMSType string

const (
	PostgreSQL RDBMSType = "postgresql"
	MySQL      RDBMSType = "mysql"
	SQLite     RDBMSType = "sqlite"
)

// ExportSQLScript generates batch INSERT SQL statement script for RDBMS
func ExportSQLScript(m *motion.MotionData, outputPath string, dbType RDBMSType, tableName string) error {
	if tableName == "" {
		tableName = "motion_frames"
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create SQL script file: %w", err)
	}
	defer file.Close()

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

	if _, err := file.WriteString(schema + "\nBEGIN TRANSACTION;\n"); err != nil {
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
		if _, err := file.WriteString(stmt); err != nil {
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

	if _, err := file.WriteString("COMMIT;\n"); err != nil {
		return err
	}

	return nil
}
