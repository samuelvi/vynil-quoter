# Reprocess Images and Upsert CSV Design

## Context

VinylQuoter processes images through `data/src → data/dst → model → CSV`. The current update path reads existing CSV rows and uses `catalog.Pending()` to skip images already present in the report. That makes normal menu and CLI runs cheap, but it prevents users from refreshing an already-processed image after changing source files, crop behavior, provider, or model.

## Goal

All normal processing entrypoints must share the same internal behavior: process the requested source images every time, overwrite their prepared image in `data/dst`, and update the CSV row identified by `source_image`. This applies to the interactive menu and CLI/Make targets.

## Accepted Behavior

- Main menu option `1` reprocesses the selected image even if its row already exists.
- Main menu option `2` reprocesses every supported image in `data/src` even if rows already exist.
- CLI equivalents such as `make run IMAGE=...`, `make run-all`, and `make run-cli ARGS="--all ..."` use the same internal path.
- `crop.Process()` remains responsible for writing `data/dst`; because it uses `os.Create`, successful decode paths overwrite existing destination images.
- CSV writes become an upsert by image ID: replace an existing row when `source_image` matches by basename, otherwise append a new row.
- Rows for images not selected in the current run are preserved when `replace=false`.
- `replace=true` still starts from an empty CSV and regenerates rows only from the selected all-images run.

## Design

Add a catalog upsert helper that compares rows with `ImageID()` on both the existing row and new row. This preserves compatibility with legacy rows that stored paths such as `data/src/DSC01.jpg` while writing new rows as basenames such as `DSC01.jpg`.

Update `app.Process()` so it no longer asks `catalog.Pending()` for missing images. Instead, it processes every image supplied by `imageinput.Collect()`. For each successful or manual-review identification result, it builds a `catalog.Row`, upserts it into the in-memory row slice, and writes the CSV after each image. Incremental writing keeps the current durability behavior if later images fail.

Remove or stop relying on the old pending-skip semantics, because they directly conflict with the new requirement. Documentation should describe normal processing as refresh/upsert, not append-missing-only.

## Testing Strategy

- Unit-test the catalog upsert helper for replacement, append, basename matching, and order preservation.
- Unit-test `Process()` with an existing CSV row to prove the recognizer is called again and the row is replaced, not duplicated.
- Unit-test destination overwrite by pre-creating a stale `dst` file and proving reprocessing replaces it.
- Update interactive-flow tests so option `2` followed by option `1` recognizes the second selected image again.
- Run `make test` and `make quality` inside the project Docker workflow.
