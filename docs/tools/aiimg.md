# aiimg

`aiimg` extracts image metadata (EXIF) from common formats and prints device, capture time, GPS coordinates, and a reverse-geocoded address. It can also **set** or **clear** writable EXIF fields on **JPEG** files (default: write a new sibling file).

## Build

```bash
make build
./bin/aiimg --help
```

## Usage

```bash
aiimg photo.jpg
aiimg --json photo.jpg shot.png
aiimg --all photo.jpg              # locked summary + full structured EXIF
aiimg --json --all photo.jpg       # locked keys + nested "exif" object
aiimg --no-geocode photo.jpg       # skip network; Location shows 0; GPS still shown when present
aiimg --version                    # or -V; also: aiimg version
aiimg version --help               # subcommand help (not: aiimg help version)

aiimg set --make Canon photo.jpg                 # writes photo_aiimg.jpg by default
aiimg set --make Canon --inplace photo.jpg       # overwrite original
aiimg set --gps 31.2304,121.4737 -o out.jpg photo.jpg
aiimg set --datetime "2020:06:07 01:30:00" photo.jpg
aiimg clear --gps --make photo.jpg
aiimg clear --all --inplace photo.jpg
```

`aiimg version` prints build version (and takes precedence over a file named `version`).

### set / clear (JPEG only)

Writable / clearable locked fields:

- Make, Model, Software, LensMake, LensModel, SerialNumber
- DateTime
- GPS (lat/lon), Altitude

Not writable: File, Format, Width, Height, Location (Address is reverse-geocode only).

`set` requires at least one field flag. `clear` requires at least one field flag or `--all`.

**Write target (BREAKING vs earlier in-place default):**

- Default (non-TTY): write a sibling `*_aiimg.jpg` (or `*_aiimg_N.jpg` on collision); leave the source unchanged
- Default (TTY): prompt once on stderr `Overwrite original file(s)? [y/N]` (`y` = overwrite; anything else = sibling file)
- `-o` / `--output path`: write to that path (single input only; may overwrite an existing file; no prompt)
- `--inplace` / `-y`: overwrite each input with no prompt
- `-o` cannot combine with multiple inputs or with `--inplace`/`-y`
- `-o` rejecting an existing directory

PNG/TIFF/WebP return unsupported.

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

By default, when GPS is present, `aiimg` calls the OpenStreetMap [Nominatim](https://nominatim.org/) reverse API (`accept-language=zh-CN,en`). This requires network access and follows Nominatim usage policy (User-Agent set; ≥1s between requests).

Use `--no-geocode` for offline use.

### Formats

JPEG and PNG are supported for **read**. TIFF/WebP are best-effort when the decoder recognizes them. **set/clear** support JPEG only.
