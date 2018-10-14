package main

// defaultConfig returns an object with the default configuration.
func defaultConfig() Config {
	var c Config

	c.Datetime.On = true
	c.Datetime.Fg = "255"
	c.Datetime.Bg = "232"

	c.Username.On = true
	c.Username.Fg = "255"
	c.Username.Bg = "235"

	c.Hostname.On = true
	c.Hostname.Fg = "255"
	c.Hostname.Bg = "238"

	c.HomeDir.Fg = "255"
	c.HomeDir.Bg = "241"

	c.RdonlyDir.Fg = "255"
	c.RdonlyDir.Bg = "124"

	c.CurrentDir.Size = 2
	c.CurrentDir.Fg = "255"
	c.CurrentDir.Bg = "244"

	c.Repository.On = true
	c.Repository.Fg = "255"
	c.Repository.Bg = "247"

	c.Symbol.Regular = "$"
	c.Symbol.SuperUser = "#"

	c.Status.Symbol = "255"
	c.Status.Success = "249"
	c.Status.Failure = "001"
	c.Status.Misuse = "003"
	c.Status.Permission = "004"
	c.Status.NotFound = "014"
	c.Status.InvalidExit = "202"
	c.Status.Terminated = "013"

	return c
}
