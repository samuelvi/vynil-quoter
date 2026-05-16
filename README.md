# VinylQuoter

VinylQuoter identifies vinyl records from images in `src/` and writes a valuation catalog to `report/album_catalog.csv`.

The default recognizer is the local LM Studio vision model `qwen2.5-vl-7b-instruct`. Gemini is also supported when `GEMINI_API_KEY` is configured.

Use it when you have photos of vinyl covers and want a quick CSV with album, artist, estimated EUR value, confidence, and review notes.

## Requirements

- Python 3.11+
- Images in `src/`
- For the default local mode: LM Studio running at `http://localhost:1234/v1` with `qwen2.5-vl-7b-instruct` loaded
- For Gemini mode: `GEMINI_API_KEY` in the environment

Supported image extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

Local cache/runtime files belong under `.cache/` in this project. The directory is ignored by Git.

## Recommended flow

1. Put source images in `src/`.
2. Start LM Studio and load `qwen2.5-vl-7b-instruct`.
3. Run `make run-all` to update the catalog.
4. Review low-confidence or `manual-review` rows in `report/album_catalog.csv`.

## Interactive usage

```bash
python3 vinyl_quoter.py
```

Menu options:

1. Process one image
2. Process all images in `src/`
3. Update the default final CSV
4. Replace/regenerate the default final CSV
5. Exit

After choosing the processing mode, choose the recognition provider:

1. LM Studio local - `qwen2.5-vl-7b-instruct` default
2. Gemini - `gemini-2.5-flash-lite`

Choose Gemini if the local model is not running or if you want a stronger cloud fallback.

## Command-line usage

```bash
# Process one image
python3 vinyl_quoter.py --image src/DSC01.jpg

# Update report/album_catalog.csv with all supported images from src/
python3 vinyl_quoter.py --all

# Replace/regenerate report/album_catalog.csv
python3 vinyl_quoter.py --all --replace

# Use Gemini instead of local LM Studio
python3 vinyl_quoter.py --all --provider gemini

# Use a custom LM Studio model or endpoint
python3 vinyl_quoter.py --all --provider lm-studio --model qwen2.5-vl-7b-instruct --lm-studio-base-url http://localhost:1234/v1
```

## Makefile commands

```bash
make help
make run
make run IMAGE=src/DSC01.jpg
make run-all
make run-all-replace
make run-gemini
make test
make quality
make clean
```

## CSV output

Default path: `report/album_catalog.csv`

Columns:

- `source_image`
- `artist`
- `title`
- `identification_confidence`
- `recommended_price_eur`
- `price_confidence`
- `price_basis`
- `notes`

Update mode keeps existing rows and only processes images that are not already present in the CSV. Replace mode regenerates the CSV from scratch.

Prices are conservative estimates for Spain/EU second-hand sales, assuming media VG+ and sleeve VG when the exact pressing is unknown.

## Troubleshooting

- `GEMINI_API_KEY is required`: choose LM Studio or export a Gemini API key.
- `LM Studio request failed`: start LM Studio, load the vision model, and confirm the local server is enabled.
- Missing rows in update mode: use `make run-all-replace` to regenerate the CSV.

## Testing

```bash
make test
make quality
```
