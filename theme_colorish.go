package main

func ApplyColorishTheme(cfg Config) Config {
	cfg.TimeOn = true
	cfg.TimeFg = 255
	cfg.TimeBg = 23
	cfg.UserOn = true
	cfg.UserFg = 255
	cfg.UserBg = 39
	cfg.HostOn = true
	cfg.HostFg = 255
	cfg.HostBg = 62
	cfg.HomeFg = 255
	cfg.HomeBg = 161
	cfg.RodirFg = 255
	cfg.RodirBg = 124
	cfg.CwdN = 2
	cfg.CwdOn = true
	cfg.CwdFg = 251
	cfg.CwdBg = 238
	cfg.RepoOn = true
	cfg.RepoFg = 0
	cfg.RepoBg = 148
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 255
	cfg.StatusSuccess = 70
	cfg.StatusError = 1
	cfg.StatusMisuse = 3
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 14
	cfg.StatusErrSignal = 202
	cfg.StatusTerminated = 13
	return cfg
}
