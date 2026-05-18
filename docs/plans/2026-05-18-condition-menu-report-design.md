# Condition Menu and Report Design

## Objective

Add user-selectable vinyl media and sleeve conditions to the interactive menu, carry those values into recognition/pricing, write the selected condition into the CSV report, improve saved reference URLs, and normalize price output so CSV values contain numbers or numeric ranges only.

## Grading Standard

Discogs uses the Goldmine Standard for Marketplace item grading. The supported grades will be:

- `M`
- `NM/M-`
- `VG+`
- `VG`
- `G+`
- `G`
- `F`
- `P`
- `Generic` for sleeve only

Default condition is `VG` for both media and sleeve because it is the requested middle/default quality.

## Recommended Approach

Use two internal config fields and one CSV column:

- `RunConfig.MediaCondition`
- `RunConfig.SleeveCondition`
- CSV column: `condition`

The CSV value will be formatted as:

```text
media: VG; sleeve: VG
```

This keeps the report compact while preserving the two user decisions separately in the app state and prompt.

## Menu Design

The main menu will show selected values in parentheses:

```text
Vinyl Quoter
1) Procesar una imagen concreta
2) Procesar todas las imágenes de data/src
3) Guardado csv (data/report/album_catalog.csv)
4) Modelo (lm-studio: qwen2.5-vl-7b-instruct)
5) Calidad carátula (VG)
6) Calidad vinilo (VG)
7) Salir
```

Selecting either condition opens a grading submenu. Sleeve condition includes `Generic`; media condition does not.

## CLI Design

Add optional flags:

```text
--media-condition VG
--sleeve-condition VG
```

Invalid values should return an argument error instead of silently accepting bad report data.

## Recognition and Price Prompt

The prompt will include selected condition:

```text
Price assumptions: Spain/EU market, EUR currency, media VG, sleeve VG, normal second-hand sale.
Return recommended_price_eur as numbers only, without currency symbols or text, e.g. "12" or "12-18".
```

The app will also sanitize provider output before writing CSV so values like `15-20 EUR`, `€12`, or `12 euros` become `15-20` or `12`.

## CSV Design

Update header from:

```text
recommended_price_eur
```

to:

```text
recommended_price_eur
condition
```

The currency remains represented by the price column name, not in cell values.

CSV reader stays backward compatible with existing rows that lack `condition` or reference URL columns.

## Reference URL Design

The current URLs include condition terms such as `VG+ sleeve VG+`, which can reduce marketplace search results. The new URLs will use cleaner, platform-appropriate queries:

- Discogs: `artist title`, `type=release`
- eBay: `artist title vinyl lp`
- Popsike: `artist title`

Unknown/empty artist and title values still produce safe generic URLs, but with no condition terms.

## Testing Strategy

Use TDD for each behavior:

1. Defaults and CLI validation.
2. Menu labels and condition selection.
3. CSV header/read/write compatibility.
4. App process attaches condition and sanitizes price.
5. Reference URL queries omit condition hints.
6. Prompt includes dynamic condition and numeric-only price contract.

All verification runs through `make test` inside Docker and the strict quick quality gate.
