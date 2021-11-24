package main

func ApplyGrayScaleTheme(cfg Config) Config {
	cfg.TimeOn = true
	cfg.TimeFg = 255
	cfg.TimeBg = 232
	cfg.UserOn = true
	cfg.UserFg = 255
	cfg.UserBg = 235
	cfg.HostOn = true
	cfg.HostFg = 255
	cfg.HostBg = 238
	cfg.HomeFg = 255
	cfg.HomeBg = 241
	cfg.RodirFg = 255
	cfg.RodirBg = 124
	cfg.CwdOn = true
	cfg.CwdFg = 255
	cfg.CwdBg = 244
	cfg.RepoOn = true
	cfg.RepoFg = 255
	cfg.RepoBg = 247
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 255
	cfg.StatusSuccess = 249
	cfg.StatusError = 1
	cfg.StatusMisuse = 3
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 14
	cfg.StatusErrSignal = 202
	cfg.StatusTerminated = 13
	return cfg
}
