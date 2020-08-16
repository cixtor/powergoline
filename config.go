package main

import (
	"fmt"
	"strings"
)

// Config represents all the available program options.
type Config struct {
	Debug            bool
	TimeOn           bool
	TimeFg           int
	TimeBg           int
	TimeFmt          string
	UserOn           bool
	UserFg           int
	UserBg           int
	HostOn           bool
	HostFg           int
	HostBg           int
	HomeFg           int
	HomeBg           int
	RodirFg          int
	RodirBg          int
	CwdN             int
	CwdOn            bool
	CwdFg            int
	CwdBg            int
	RepoOn           bool
	RepoFg           int
	RepoBg           int
	RepoExclude      FlagStringArray
	RepoInclude      FlagStringArray
	Plugins          FlagPluginArray
	SymbolRoot       string
	SymbolUser       string
	StatusFg         int
	StatusCode       int
	StatusSuccess    int
	StatusError      int
	StatusMisuse     int
	StatusCantExec   int
	StatusNotFound   int
	StatusInvalid    int
	StatusErrSignal  int
	StatusTerminated int
	StatusOutofrange int
}

type FlagStringArray []string

func (v FlagStringArray) Set(s string) error {
	return nil
}

func (v FlagStringArray) String() string {
	return ""
}

type FlagPluginArray []Plugin

func (v *FlagPluginArray) Set(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("invalid plugin")
	}
	pieces := strings.Split(s, "\x20")
	*v = append(*v, Plugin{
		Name: pieces[0],
		Args: pieces[1:],
	})
	return nil
}

func (v FlagPluginArray) String() string {
	return ""
}

type Plugin struct {
	Name string
	Args []string
}
