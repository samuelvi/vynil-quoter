# VinylQuoter CSV Output

VinylQuoter writes its default report to:

```text
data/report/album_catalog.csv
```

The CSV is updated throughout processing so partial results are preserved if a later image fails.

## Columns

| Column | Meaning |
| --- | --- |
| `source_image` | Original input file name used as the stable row ID, for example `DSC01.jpg`. |
| `artist` | Recognized artist or band. |
| `title` | Recognized album title. |
| `identification_confidence` | Provider confidence or `manual-review` when recognition fails. |
| `recommended_price_eur` | Conservative EUR value as numbers only, such as `12` or `12-18`. |
| `condition` | Selected grades as `media: <grade>; sleeve: <grade>`. |
| `price_confidence` | Provider confidence for the price estimate. |
| `price_basis` | Short explanation of the valuation basis. |
| `notes` | Review notes, warnings, or failure details. |
| `discogs_reference_url` | Broad Discogs release search from artist/title. |
| `ebay_reference_url` | Broad eBay vinyl/LP search from artist/title. |
| `popsike_reference_url` | Broad Popsike search from artist/title. |

## Update and replace behavior

Normal processing updates rows by `source_image`:

- Selected images are reprocessed every run.
- Existing rows with the same `source_image` are replaced with fresh results.
- Rows for images not selected in the current run are preserved.
- Single-image runs always update/create rows and never replace the whole CSV.

Replace mode rebuilds the CSV from scratch for all-images runs:

```bash
make run-all-replace
```

## Prices

`recommended_price_eur` stores only numeric values because the currency is already expressed in the column name.

Examples:

```text
12
12-18
7.5
```

If the provider returns text such as `about €12 to €18`, the app extracts the first one or two numeric values and writes `12-18`.

## Conditions

The selected media and sleeve grades affect the provider prompt and are persisted in the `condition` column.

Media grades:

```text
M, NM/M-, VG+, VG, G+, G, F, P
```

Sleeve grades:

```text
M, NM/M-, VG+, VG, G+, G, F, P, Generic
```

Default:

```text
media: VG; sleeve: VG
```

## Reference URLs

Reference URLs are intentionally broad searches based on artist and title. They do not include condition terms such as `VG+` or `sleeve`, because condition-heavy searches often hide useful comparable listings.

If artist or title is missing or `Unknown`, it is omitted from the search query.
