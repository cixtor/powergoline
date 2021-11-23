package main

func ApplyBlueScaleTheme(cfg Config) Config {
	cfg.TimeOn = true
	cfg.TimeFg = 255
	cfg.TimeBg = 18
	cfg.UserOn = true
	cfg.UserFg = 255
	cfg.UserBg = 31
	cfg.HostOn = true
	cfg.HostFg = 255
	cfg.HostBg = 38
	cfg.HomeFg = 81
	cfg.HomeBg = 31
	cfg.RodirFg = 255
	cfg.RodirBg = 54
	cfg.CwdN = 4
	cfg.CwdOn = true
	cfg.CwdFg = 81
	cfg.CwdBg = 24
	cfg.RepoOn = true
	cfg.RepoFg = 255
	cfg.RepoBg = 75
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 0
	cfg.StatusSuccess = 255
	cfg.StatusError = 162
	cfg.StatusMisuse = 3
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 14
	cfg.StatusErrSignal = 202
	cfg.StatusTerminated = 240
	return cfg
}
