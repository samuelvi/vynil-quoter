# VinylQuoter Quickstart

1. Put vinyl cover images in `data/src/`.
2. Build and start the test container:
   ```bash
   make test-build
   make test-up
   ```
3. Generate or update the catalog:
   ```bash
   make run-all
   ```
4. Review `data/report/album_catalog.csv`.
5. Stop the container:
   ```bash
   make test-down
   ```

Use Gemini instead of LM Studio with:

```bash
make run-gemini
```
