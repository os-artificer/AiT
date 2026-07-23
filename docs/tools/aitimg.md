# aitimg

`aitimg` extracts image metadata (EXIF) from common formats and prints device, capture time, GPS coordinates, and a reverse-geocoded address. It can also **set** or **clear** writable EXIF fields on **JPEG** files in place.

## Build

```bash
make build
./bin/aitimg --help
```

## Usage

```bash
aitimg photo.jpg
aitimg --json photo.jpg shot.png
aitimg --all photo.jpg              # locked summary + full structured EXIF
aitimg --json --all photo.jpg       # locked keys + nested "exif" object
aitimg --no-geocode photo.jpg       # skip network; Location shows 0; GPS still shown when present
aitimg --version                    # or -V; also: aitimg version
aitimg version --help               # subcommand help (not: aitimg help version)

aitimg set --make Canon photo.jpg
aitimg set --gps 31.2304,121.4737 --altitude 12 photo.jpg
aitimg set --datetime "2020:06:07 01:30:00" photo.jpg
aitimg clear --gps --make photo.jpg
aitimg clear --all photo.jpg
```

`aitimg version` prints build version (and takes precedence over a file named `version`).

### set / clear (JPEG only)

Writable / clearable locked fields:

- Make, Model, Software, LensMake, LensModel, SerialNumber
- DateTime
- GPS (lat/lon), Altitude

Not writable: File, Format, Width, Height, Location (Address is reverse-geocode only).

`set` requires at least one field flag. `clear` requires at least one field flag or `--all`. Writes are in place (temp file + rename). PNG/TIFF/WebP return unsupported.

### Output

Human and JSON always print every locked field. Missing values show as `0` (GPS as `0.000000, 0.000000`; Altitude as `0 m` / JSON `0`):

- File, Format, Width, Height
- Make, Model, Software, LensMake, LensModel, SerialNumber
- DateTime
- Location (Nominatim address when geocode succeeds; otherwise `0`)
- GPS (EXIF coordinates, or zeros when absent)
- Altitude (when present, else `0`)

JSON mode always emits a **JSON array** (one object per successful file) with the same locked keys.

#### `--all` (structured EXIF)

After the locked summary (or under `"exif"` with `--json`), prints the TIFF/EXIF IFD tree as parsed by `imagemeta`:

- Human mode: each section is a two-column **Field / Value** table (aligned columns)
- Root: ImageType, CameraSerial, CameraMakeID (when present)
- Sections in order: **IFD0 → ExifIFD → GPS → IFD1 → IFD2 → MakerNote** (omit empty/null sections)
- Includes exposure/optics fields (ExposureTime, FNumber, ISO, FocalLength, …) when present
- Skips absent/empty values in the full dump (unlike the locked summary’s always-`0` policy)
- **String fields** (e.g. `UserComment`): auto-detect **JSON**, **XML**, or **YAML**. Structured documents flatten to dotted paths (`UserComment.AIGC.Label`); otherwise show the **raw** string. Nested structured strings expand up to depth 3. Long values wrap onto continuation lines (never truncated)
- **Not** a raw APP1 tag dump; XMP/ICC/IPTC and unmapped private tags are out of scope
- `--json --all` still emits nested `"exif"` machine JSON (unchanged shape)

### Network

By default, when GPS is present, `aitimg` calls the OpenStreetMap [Nominatim](https://nominatim.org/) reverse API (`accept-language=zh-CN,en`). This requires network access and follows Nominatim usage policy (User-Agent set; ≥1s between requests).

Use `--no-geocode` for offline use.

### Formats

JPEG and PNG are supported for **read**. TIFF/WebP are best-effort when the decoder recognizes them. **set/clear** support JPEG only.
