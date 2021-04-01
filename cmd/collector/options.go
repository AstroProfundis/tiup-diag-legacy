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
	ClusterName  string `json:"cluster" long:"cluster-name" description:"specify the cluster name"`
	User         string `json:"user" long:"user" description:"the user with root permission"`
	Passwd       string `json:"pass" long:"passwd" description:"password of the user"`
	IdentityFile string `json:"key" long:"identity-file" description:"identify file for SSH login"`
}

func (o *Options) GetHome() string {
	return o.Home
}

func (o *Options) GetInspectionId() string {
	return o.InspectionID
}

func (o *Options) GetClusterName() string {
	return o.ClusterName
}

func (o *Options) GetUser() string {
	return o.User
}

func (o *Options) GetPasswd() string {
	return o.Passwd
}

func (o *Options) GetIdentityFile() string {
	return o.IdentityFile
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
