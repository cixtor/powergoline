package main

func ApplyAstrocomTheme(cfg Config) Config {
	cfg.TimeOn = true
	cfg.TimeFg = 255
	cfg.TimeBg = 23
	cfg.UserOn = true
	cfg.UserFg = 236
	cfg.UserBg = 203
	cfg.HostOn = true
	cfg.HostFg = 236
	cfg.HostBg = 208
	cfg.HomeFg = 236
	cfg.HomeBg = 226
	cfg.RodirFg = 255
	cfg.RodirBg = 124
	cfg.CwdOn = true
	cfg.CwdFg = 251
	cfg.CwdBg = 238
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 255
	cfg.StatusSuccess = 99
	cfg.StatusError = 1
	cfg.StatusMisuse = 3
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 14
	cfg.StatusErrSignal = 202
	cfg.StatusTerminated = 240
	return cfg
}
