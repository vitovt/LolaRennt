# v1.0 Backlog

Legend:
- `[ ]` pending
- `[x]` done

Current focus:
- [x] Implement real preview playback transport: `Play`, `Pause`, `Stop`, `Restart`, `Loop`, and timeline auto-advance
- [x] Make animation mode actually affect rendering for `Scramble basic`, `Scramble with lock`, and `Reveal then scramble then lock`
- [x] Make `Random switch rate` affect animation output instead of being UI-only
- [x] Add preview controls: zoom, checkerboard, safe area toggle, and background preview modes
- [x] Add missing layout control for `padding`
- [x] Add text validation UX: supported charset viewer, unsupported character highlighting, and replacement guidance
- [x] Add background image `opacity`
- [x] Make `supersampling` affect export output instead of only being stored in project state

## P0 Critical

## P1 Important

- [x] Add export `ETA`
- [x] Implement `Open output folder`
- [x] Add `preview region / full canvas` control in export
- [x] Support manual `ffmpeg` / `ffprobe` path override
- [x] Add preset management workflow: save, duplicate, delete
- [x] Persist presets as JSON instead of shipping only hardcoded presets

## P2 Nice To Have Within Spec

- [x] Add `flicker` control
- [x] Add `noise overlay`
- [x] Add `scanline` toggle
- [x] Add `simultaneous final lock`
- [x] Add `immediate punctuation lock`
