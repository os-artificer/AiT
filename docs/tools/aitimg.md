# aitimg

`aitimg` extracts image metadata (EXIF) from common formats and prints device, capture time, GPS coordinates, and a reverse-geocoded address.

## Build

```bash
make build
./bin/aitimg --help
```

## Usage

```bash
aitimg photo.jpg
aitimg --json photo.jpg shot.png
aitimg --no-geocode photo.jpg   # skip network; Location shows 0; GPS still shown when present
aitimg --version
```

### Output

Human and JSON always print every locked field. Missing values show as `0` (GPS as `0.000000, 0.000000`; Altitude as `0 m` / JSON `0`):

- File, Format, Width, Height
- Make, Model, Software, LensMake, LensModel, SerialNumber
- DateTime
- Location (Nominatim address when geocode succeeds; otherwise `0`)
- GPS (EXIF coordinates, or zeros when absent)
- Altitude (when present, else `0`)

JSON mode always emits a **JSON array** (one object per successful file) with the same locked keys.

### Network

By default, when GPS is present, `aitimg` calls the OpenStreetMap [Nominatim](https://nominatim.org/) reverse API (`accept-language=zh-CN,en`). This requires network access and follows Nominatim usage policy (User-Agent set; ≥1s between requests).

Use `--no-geocode` for offline use.

### Formats

JPEG and PNG are supported. TIFF/WebP are best-effort when the decoder recognizes them.
