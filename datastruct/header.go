package datastruct

// CleanedHeader is a structure holding header information of a replay file derived from s2prot.Rep.Header
type CleanedHeader struct {
	ElapsedGameLoops    uint64  `json:"elapsedGameLoops"`
	DurationNanoseconds int64   `json:"durationNanoseconds"`
	DurationSeconds     float64 `json:"durationSeconds"`
	Version             string  `json:"version"` //s2prot.Struct
	// UseScaledTime    bool          `json:"useScaledTime"`
}
