# v1.0 Backlog

Legend:
- `[ ]` pending
- `[x]` done

Current focus:
- [ ] Implement real preview playback transport: `Play`, `Pause`, `Stop`, `Restart`, `Loop`, and timeline auto-advance

## P0 Critical

- [ ] Make animation mode actually affect rendering for `Scramble basic`, `Scramble with lock`, and `Reveal then scramble then lock`
- [ ] Make `Random switch rate` affect animation output instead of being UI-only
- [ ] Add preview controls: zoom, checkerboard, safe area toggle, and background preview modes
- [ ] Add text validation UX: supported charset viewer, unsupported character highlighting, and replacement guidance
- [ ] Add missing layout control for `padding`
- [ ] Add background image `opacity`
- [ ] Make `supersampling` affect export output instead of only being stored in project state

## P1 Important

- [ ] Add export `ETA`
- [ ] Implement `Open output folder`
- [ ] Add `preview region / full canvas` control in export
- [ ] Support manual `ffmpeg` / `ffprobe` path override
- [ ] Add preset management workflow: save, duplicate, delete
- [ ] Persist presets as JSON instead of shipping only hardcoded presets

## P2 Nice To Have Within Spec

- [ ] Add `flicker` control
- [ ] Add `noise overlay`
- [ ] Add `scanline` toggle
- [ ] Add `simultaneous final lock`
- [ ] Add `immediate punctuation lock`
