package main

const white = "255"

// defaultConfig returns an object with the default configuration.
func defaultConfig() Config {
	var c Config

	c.Datetime.On = false
	c.Datetime.Fg = white
	c.Datetime.Bg = "013"

	c.Username.On = false
	c.Username.Fg = white
	c.Username.Bg = "033"

	c.Hostname.On = false
	c.Hostname.Fg = white
	c.Hostname.Bg = "105"

	c.HomeDir.Fg = white
	c.HomeDir.Bg = "093"

	c.RdonlyDir.Fg = white
	c.RdonlyDir.Bg = "124"

	c.CurrentDir.Size = 1
	c.CurrentDir.Fg = white
	c.CurrentDir.Bg = "099"

	c.Repository.On = false
	c.Repository.Fg = "000"
	c.Repository.Bg = "148"

	c.Symbol.Regular = "$"
	c.Symbol.SuperUser = "#"

	c.Status.Symbol = white
	c.Status.Success = "041"
	c.Status.Failure = "001"
	c.Status.Misuse = "003"
	c.Status.Permission = "004"
	c.Status.NotFound = "014"
	c.Status.InvalidExit = "202"
	c.Status.Terminated = "013"

	return c
}
