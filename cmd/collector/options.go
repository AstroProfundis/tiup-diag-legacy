package main

import (
	"errors"
	"time"
)

type Options struct {
	Home         string `json:"home" long:"home" description:"foresight working directory" required:"true"`
	Begin        string `json:"begin" long:"begin" description:"scrape begin time" required:"true"`
	End          string `json:"end" long:"end" description:"scrape begin time" required:"true"`
	InspectionID string `json:"id" long:"id" description:"specify an ID of the operation"`
}

func (o *Options) GetHome() string {
	return o.Home
}

func (o *Options) GetInspectionId() string {
	return o.InspectionID
}

func (o *Options) GetScrapeBegin() (time.Time, error) {
	if o.Begin == "" {
		return time.Time{}, errors.New("begin time not specified in command line")
	}
	return time.Parse(time.RFC3339, o.Begin)
}

func (o *Options) GetScrapeEnd() (time.Time, error) {
	if o.End == "" {
		return time.Time{}, errors.New("end time not specified in command line")
	}
	return time.Parse(time.RFC3339, o.End)
}
