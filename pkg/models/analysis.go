package models

import "time"

type Analysis struct {
	Action     string
	Confidence float64
	Reasoning  string
	Model      string
	Timestamp  time.Time
}
