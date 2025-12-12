# Image Metadata & Rule-Checker CLI Tool

A command-line tool written in Go that processes images, extracts EXIF metadata and image internals, computes SHA256 hashes, and applies configurable rule-based checks. Outputs structured JSON results.

## Features

- **EXIF Metadata Extraction**: Parses camera maker, timestamp, GPS coordinates, and orientation
- **Image Format Parsing**: Binary parsing of JPEG (entropy segments, dimensions, color model) and PNG (chunks, dimensions, color model)
- **Hash Computation**: SHA256 hash of each image file
- **Rule Engine**: Configurable rule-based checks (dimensions, timestamp, GPS, orientation)
- **Structured Output**: JSON output with detailed results per image
- **Logging**: Configurable log levels (debug, info, warn, error)

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/k3rne1pan1c/image-checker.git
cd image-checker

# Build the binary
make build

# Or use go directly
go build -o bin/imgcheck ./cmd/imgcheck
```

### Docker

```bash
# Build Docker image
make docker-build

# Or manually
docker build -t imgcheck:latest .
```

## Usage

### Basic Usage

```bash
# Process images in a folder
./bin/imgcheck /path/to/images

# Output to file
./bin/imgcheck -output results.json /path/to/images

# Use custom rule configuration
./bin/imgcheck -config custom-rules.yaml /path/to/images

# Recursive processing
./bin/imgcheck -recursive /path/to/images

# Debug logging
./bin/imgcheck -log-level debug /path/to/images
```

### Command-Line Options

- `-config string`: Path to rule configuration file (default: `config/rules.yaml`)
- `-output string`: Path to output JSON file (default: stdout)
- `-log-level string`: Log level: debug, info, warn, error (default: `info`)
- `-recursive`: Process images recursively in subdirectories

### Docker Usage

```bash
# Run in Docker
docker run --rm -v /path/to/images:/images imgcheck:latest /images

# With custom config
docker run --rm -v /path/to/images:/images -v /path/to/config:/app/config imgcheck:latest -config /app/config/rules.yaml /images
```

## Configuration

Rules are configured via a YAML file. See `config/rules.yaml` for the default configuration.

### Rule Configuration Format

```yaml
rules:
  min_dimensions:
    enabled: true
    width: 800
    height: 600
  max_dimensions:
    enabled: false
    width: 10000
    height: 10000
  require_timestamp:
    enabled: true
  gps_allowed:
    enabled: true
    allowed: false  # GPS not allowed
  require_orientation:
    enabled: false
```

### Available Rules

- **min_dimensions**: Check minimum image width and height
- **max_dimensions**: Check maximum image width and height
- **require_timestamp**: Ensure EXIF timestamp exists
- **gps_allowed**: Check if GPS data is allowed/not allowed
- **require_orientation**: Check if orientation tag exists

## Output Format

The tool outputs JSON with the following structure:

```json
{
  "processed": 5,
  "passed": 3,
  "failed": 2,
  "images": [
    {
      "path": "photo1.jpg",
      "hash": "abc123...",
      "exif": {
        "maker": "Canon",
        "timestamp": "2024-01-15T10:30:00Z",
        "gps": {"lat": 37.7749, "lon": -122.4194},
        "orientation": 1
      },
      "image": {
        "dimensions": {"width": 1920, "height": 1080},
        "color_model": "RGB",
        "compression": "JPEG"
      },
      "rules": [
        {"name": "min_dimensions", "passed": true},
        {"name": "require_timestamp", "passed": true},
        {"name": "gps_allowed", "passed": false, "message": "GPS not allowed but present"}
      ],
      "status": "failed"
    }
  ]
}
```

See `output/sample-output.json` for a complete example.

## Makefile Targets

- `make build`: Build the binary
- `make test`: Run tests
- `make run FOLDER=/path/to/images`: Run the application
- `make docker-build`: Build Docker image
- `make docker-run FOLDER=/path/to/images`: Run Docker container
- `make clean`: Clean build artifacts
- `make deps`: Download and tidy dependencies
- `make fmt`: Format code
- `make lint`: Lint code
- `make help`: Show help message

## Architecture

The tool is organized into the following components:

- **Parser**: EXIF, JPEG, and PNG parsing modules
- **Hash**: SHA256 computation
- **Rules**: Configurable rule engine
- **Model**: Data structures for results
- **CLI**: Command-line interface and orchestration

## Supported Image Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)

## Dependencies

- `github.com/rwcarlsen/goexif/exif`: EXIF parsing
- `gopkg.in/yaml.v3`: YAML configuration parsing
- Standard library: `crypto/sha256`, `encoding/json`, `log/slog`

## Development

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Build
make build
```

## License

MIT

