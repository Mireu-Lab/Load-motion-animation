import argparse
import sys
import os
from .parser import load_motion, iter_motion_frames
from .dataframe import to_dataframe
from .exporter import export_csv, export_sql, export_json, export_stream_csv, export_stream_sql
from .visualizer import export_gif

def main():
    parser = argparse.ArgumentParser(description="Load-Motion-Animation CLI (Python)")
    parser.add_argument("-input", required=True, help="Path to motion file (.bvh, .c3d, .trc)")
    parser.add_argument("-export", choices=["summary", "csv", "tsv", "sql", "json", "gif"], default="summary", help="Export format")
    parser.add_argument("-output", help="Output file path")
    parser.add_argument("-db", choices=["sqlite", "postgresql", "mysql"], default="sqlite", help="RDBMS dialect for SQL export")
    parser.add_argument("-max-frames", type=int, default=120, help="Maximum frames for GIF rendering")
    parser.add_argument("-stream", action="store_true", help="Enable low-memory streaming mode")

    args = parser.parse_args()

    if args.stream and args.export in ("csv", "tsv", "sql"):
        print(f"⚡ Streaming motion file: {args.input} ...")
        frames_iter = iter_motion_frames(args.input)
        if args.export == "csv":
            out = args.output or (os.path.splitext(args.input)[0] + ".csv")
            export_stream_csv(frames_iter, out, delimiter=",")
            print(f"💾 Streamed CSV to: {out}")
            return
        elif args.export == "tsv":
            out = args.output or (os.path.splitext(args.input)[0] + ".tsv")
            export_stream_csv(frames_iter, out, delimiter="\t")
            print(f"💾 Streamed TSV to: {out}")
            return
        elif args.export == "sql":
            out = args.output or (os.path.splitext(args.input)[0] + ".sql")
            export_stream_sql(frames_iter, out, db_type=args.db, file_name=args.input)
            print(f"💾 Streamed SQL ({args.db}) script to: {out}")
            return

    print(f"📂 Loading motion file: {args.input} ...")
    try:
        motion_data = load_motion(args.input)
    except Exception as e:
        print(f"❌ Error loading file: {e}")
        sys.exit(1)

    print(f"✅ Parsed successfully! Format: {motion_data.format}, Total Frames: {motion_data.total_frames}, Frame Rate: {motion_data.frame_rate:.2f} FPS, Joint/Marker Count: {len(motion_data.joint_list)}")

    ext = os.path.splitext(args.input)[1].lower()

    if args.export == "summary":
        df_dict = to_dataframe(motion_data, library="dict")
        num_rows = len(df_dict["frame"])
        num_cols = len(df_dict)
        print("📊 Data Summary:")
        print(f"DataFrame: {num_rows} rows x {num_cols} columns")
        if motion_data.joint_list:
            print(f"🦴 Joints/Markers ({len(motion_data.joint_list)} total): {motion_data.joint_list[:10]}")

    elif args.export == "csv":
        out = args.output or (os.path.splitext(args.input)[0] + ".csv")
        export_csv(motion_data, out, delimiter=",")
        print(f"💾 Exported CSV to: {out}")

    elif args.export == "tsv":
        out = args.output or (os.path.splitext(args.input)[0] + ".tsv")
        export_csv(motion_data, out, delimiter="\t")
        print(f"💾 Exported TSV to: {out}")

    elif args.export == "sql":
        out = args.output or (os.path.splitext(args.input)[0] + ".sql")
        export_sql(motion_data, out, db_type=args.db)
        print(f"💾 Exported SQL ({args.db}) script to: {out}")

    elif args.export == "json":
        out = args.output or (os.path.splitext(args.input)[0] + ".json")
        export_json(motion_data, out)
        print(f"💾 Exported JSON document to: {out}")

    elif args.export == "gif":
        out = args.output or (os.path.splitext(args.input)[0] + ".gif")
        export_gif(motion_data, out, max_frames=args.max_frames)
        print(f"🎬 Rendered animated GIF to: {out}")

if __name__ == "__main__":
    main()
