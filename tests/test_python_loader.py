import unittest
import os
import sys
import tempfile

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "python")))

from load_motion import load_motion, to_dataframe, export_csv, export_sql, export_json

class TestPythonLoader(unittest.TestCase):
    def setUp(self):
        self.sample_bvh = os.path.join(os.path.dirname(__file__), "..", "_sample", "Sample_motion.bvh")
        self.sample_c3d = os.path.join(os.path.dirname(__file__), "..", "_sample", "arthuman-sample.c3d")

    def test_bvh_loading(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        data = load_motion(self.sample_bvh)
        self.assertEqual(data.format, "BVH")
        self.assertGreater(data.total_frames, 0)
        self.assertGreater(len(data.joint_list), 0)

    def test_c3d_loading(self):
        if not os.path.exists(self.sample_c3d):
            self.skipTest("Sample C3D file missing")
        data = load_motion(self.sample_c3d)
        self.assertEqual(data.format, "C3D")
        self.assertGreater(data.total_frames, 0)

    def test_exporters(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        data = load_motion(self.sample_bvh)
        with tempfile.TemporaryDirectory() as tmp_dir:
            csv_file = os.path.join(tmp_dir, "out.csv")
            sql_file = os.path.join(tmp_dir, "out.sql")
            json_file = os.path.join(tmp_dir, "out.json")

            export_csv(data, csv_file)
            self.assertTrue(os.path.exists(csv_file))

            export_sql(data, sql_file, db_type="postgresql")
            self.assertTrue(os.path.exists(sql_file))

            export_json(data, json_file)
            self.assertTrue(os.path.exists(json_file))

if __name__ == "__main__":
    unittest.main()
