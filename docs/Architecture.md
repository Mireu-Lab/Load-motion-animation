# Architecture & Pipeline

```text
Raw Motion File (.bvh, .c3d, .trc)
          │
          ▼
Go High-Speed Parser (pkg/bvh, pkg/c3d, pkg/trc)
          │
          ▼
Unified Motion Representation (pkg/motion)
          │
          ├─────────────────────────┬─────────────────────────┐
          ▼                         ▼                         ▼
Dataframe Engine (pkg/df)    Exporters (CSV, DB)     Python C-Bridge (load_motion)
          │
          ▼
GIF Visualizer (pkg/visualizer)
```
