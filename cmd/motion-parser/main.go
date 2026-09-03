// motion-parser is a high-performance command-line utility for parsing, converting,
// and visualizing 3D/2D motion capture files (.bvh, .c3d, .trc).
//
// Usage:
//
//	motion-parser -input <file_path> [-export summary|csv|tsv|sql|json|gif] [-output <output_path>]
//
// Options:
//
//	-input       Path to motion capture file (.bvh, .c3d, .trc)
//	-export      Export target: summary, csv, tsv, sql, json, gif (default: summary)
//	-output      Target output file path
//	-max-frames  Maximum frames to render when exporting to GIF (default: 120)
//	-db          RDBMS SQL dialect for table DDL and insert statements: sqlite, postgresql, mysql (default: sqlite)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/bvh"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/c3d"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/df"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/exporter"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/motion"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/trc"
	"gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/visualizer"
)

func main() {
	inputFile := flag.String("input", "", "Path to motion capture file (.bvh, .c3d, .trc)")
	outputFormat := flag.String("export", "summary", "Export format: summary, csv, tsv, sql, json, gif")
	outputFile := flag.String("output", "", "Output file path (default: input file name with extension)")
	maxFrames := flag.Int("max-frames", 120, "Maximum frames for GIF rendering")
	dbType := flag.String("db", "sqlite", "RDBMS dialect for SQL export: sqlite, postgresql, mysql")
	streamMode := flag.Bool("stream", false, "Enable low-memory streaming mode for large files (csv, tsv, sql, json)")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: motion-parser -input <file_path> [-export csv|tsv|sql|json|gif|summary] [-output <output_path>] [-stream]")
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(*inputFile))
	var parser motion.Parser

	switch ext {
	case ".bvh":
		parser = bvh.NewParser()
	case ".c3d":
		parser = c3d.NewParser()
	case ".trc":
		parser = trc.NewParser()
	default:
		log.Fatalf("Unsupported file extension: %s. Supported: .bvh, .c3d, .trc", ext)
	}

	if *streamMode {
		streamParser, ok := parser.(motion.StreamParser)
		if !ok {
			log.Fatalf("Parser for %s does not support streaming mode", ext)
		}

		fmt.Printf("⚡ Running in low-memory streaming mode: %s ...\n", *inputFile)
		switch *outputFormat {
		case "csv":
			out := *outputFile
			if out == "" {
				out = strings.TrimSuffix(*inputFile, ext) + ".csv"
			}
			meta, err := exporter.ExportStreamCSV(streamParser, *inputFile, out)
			if err != nil {
				log.Fatalf("❌ Failed to stream export CSV: %v", err)
			}
			fmt.Printf("💾 Streamed CSV to: %s (Total Frames: %d)\n", out, meta.TotalFrames)
			return
		case "tsv":
			out := *outputFile
			if out == "" {
				out = strings.TrimSuffix(*inputFile, ext) + ".tsv"
			}
			meta, err := exporter.ExportStreamTSV(streamParser, *inputFile, out)
			if err != nil {
				log.Fatalf("❌ Failed to stream export TSV: %v", err)
			}
			fmt.Printf("💾 Streamed TSV to: %s (Total Frames: %d)\n", out, meta.TotalFrames)
			return
		case "sql":
			out := *outputFile
			if out == "" {
				out = strings.TrimSuffix(*inputFile, ext) + ".sql"
			}
			rdbType := exporter.RDBMSType(*dbType)
			meta, err := exporter.ExportStreamSQLScript(streamParser, *inputFile, out, rdbType, "motion_frames")
			if err != nil {
				log.Fatalf("❌ Failed to stream export SQL: %v", err)
			}
			fmt.Printf("💾 Streamed SQL script (%s) to: %s (Total Frames: %d)\n", *dbType, out, meta.TotalFrames)
			return
		case "json", "ndjson":
			out := *outputFile
			if out == "" {
				out = strings.TrimSuffix(*inputFile, ext) + ".ndjson"
			}
			meta, err := exporter.ExportStreamMongoLines(streamParser, *inputFile, out)
			if err != nil {
				log.Fatalf("❌ Failed to stream export NDJSON: %v", err)
			}
			fmt.Printf("💾 Streamed NDJSON lines to: %s (Total Frames: %d)\n", out, meta.TotalFrames)
			return
		default:
			fmt.Println("⚠️ Streaming mode only supports csv, tsv, sql, and json/ndjson exports. Falling back to in-memory mode...")
		}
	}

	fmt.Printf("📂 Loading motion file: %s ...\n", *inputFile)
	data, err := parser.ParseFile(*inputFile)
	if err != nil {
		log.Fatalf("❌ Error parsing file: %v", err)
	}

	fmt.Printf("✅ Parsed successfully! Format: %s, Total Frames: %d, Frame Rate: %.2f FPS, Joint/Marker Count: %d\n",
		data.Format, data.TotalFrames, data.FrameRate, len(data.JointList))

	switch *outputFormat {
	case "summary":
		dataframe := df.FromMotionData(data)
		fmt.Println("📊 Data Summary:")
		fmt.Println(dataframe.Summary())
		if len(data.JointList) > 0 {
			fmt.Printf("🦴 Joints/Markers (%d total): %v\n", len(data.JointList), data.JointList[:min(10, len(data.JointList))])
		}

	case "csv":
		out := *outputFile
		if out == "" {
			out = strings.TrimSuffix(*inputFile, ext) + ".csv"
		}
		if err := exporter.ExportCSV(data, out); err != nil {
			log.Fatalf("❌ Failed to export CSV: %v", err)
		}
		fmt.Printf("💾 Exported CSV to: %s\n", out)

	case "tsv":
		out := *outputFile
		if out == "" {
			out = strings.TrimSuffix(*inputFile, ext) + ".tsv"
		}
		if err := exporter.ExportTSV(data, out); err != nil {
			log.Fatalf("❌ Failed to export TSV: %v", err)
		}
		fmt.Printf("💾 Exported TSV to: %s\n", out)

	case "sql":
		out := *outputFile
		if out == "" {
			out = strings.TrimSuffix(*inputFile, ext) + ".sql"
		}
		rdbType := exporter.RDBMSType(*dbType)
		if err := exporter.ExportSQLScript(data, out, rdbType, "motion_frames"); err != nil {
			log.Fatalf("❌ Failed to export SQL script: %v", err)
		}
		fmt.Printf("💾 Exported SQL script (%s) to: %s\n", *dbType, out)

	case "json":
		out := *outputFile
		if out == "" {
			out = strings.TrimSuffix(*inputFile, ext) + ".json"
		}
		if err := exporter.ExportJSON(data, out); err != nil {
			log.Fatalf("❌ Failed to export JSON: %v", err)
		}
		fmt.Printf("💾 Exported JSON document to: %s\n", out)

	case "gif":
		out := *outputFile
		if out == "" {
			out = strings.TrimSuffix(*inputFile, ext) + ".gif"
		}
		opts := visualizer.DefaultGIFOptions()
		opts.MaxFrames = *maxFrames
		if data.FrameRate > 0 {
			opts.FPS = int(data.FrameRate)
		}
		if err := visualizer.ExportGIF(data, out, opts); err != nil {
			log.Fatalf("❌ Failed to export GIF: %v", err)
		}
		fmt.Printf("🎬 Rendered animated GIF to: %s\n", out)

	default:
		log.Fatalf("Unknown export format: %s", *outputFormat)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
