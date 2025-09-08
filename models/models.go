package models

import "time"

type Sale struct {
	ID       int       `json:"id"`
	Type     string    `json:"type" validate:"required,oneof=income expense"`
	Amount   float64   `json:"amount" validate:"required,gt=0"`
	Date     time.Time `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Category string    `json:"category" validate:"required"`
}

type AnalyticsResponse struct {
	Sum          float64 `json:"sum"`
	Average      float64 `json:"average"`
	Count        int     `json:"count"`
	Median       float64 `json:"median"`
	Percentile90 float64 `json:"percentile90"`
}

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
}
