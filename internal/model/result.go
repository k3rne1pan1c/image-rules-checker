package model

// Result represents the overall processing result
type Result struct {
	Processed int     `json:"processed"`
	Passed    int     `json:"passed"`
	Failed    int     `json:"failed"`
	Images    []Image `json:"images"`
}

// Image represents the analysis result for a single image
type Image struct {
	Path   string      `json:"path"`
	Hash   string      `json:"hash"`
	EXIF   *EXIFData   `json:"exif,omitempty"`
	Image  *ImageData  `json:"image,omitempty"`
	Rules  []RuleCheck `json:"rules"`
	Status string      `json:"status"` // "passed" or "failed"
}

// EXIFData contains extracted EXIF metadata
type EXIFData struct {
	Maker      string    `json:"maker,omitempty"`
	Timestamp  string    `json:"timestamp,omitempty"`
	GPS        *GPSData  `json:"gps,omitempty"`
	Orientation int      `json:"orientation,omitempty"`
}

// GPSData contains GPS coordinates
type GPSData struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// ImageData contains image internals
type ImageData struct {
	Dimensions  Dimensions `json:"dimensions"`
	ColorModel  string     `json:"color_model"`
	Compression string     `json:"compression"`
}

// Dimensions contains image width and height
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RuleCheck represents the result of a single rule check
type RuleCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

