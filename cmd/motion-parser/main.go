package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"load-motion-animation/pkg/bvh"
	"load-motion-animation/pkg/c3d"
	"load-motion-animation/pkg/df"
	"load-motion-animation/pkg/exporter"
	"load-motion-animation/pkg/motion"
	"load-motion-animation/pkg/trc"
	"load-motion-animation/pkg/visualizer"
)

func main() {
	inputFile := flag.String("input", "", "Path to motion capture file (.bvh, .c3d, .trc)")
	outputFormat := flag.String("export", "summary", "Export format: summary, csv, tsv, sql, json, gif")
	outputFile := flag.String("output", "", "Output file path (default: input file name with extension)")
	maxFrames := flag.Int("max-frames", 120, "Maximum frames for GIF rendering")
	dbType := flag.String("db", "sqlite", "RDBMS dialect for SQL export: sqlite, postgresql, mysql")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: motion-parser -input <file_path> [-export csv|tsv|sql|json|gif|summary] [-output <output_path>]")
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
