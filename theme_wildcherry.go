package main

func ApplyWildCherryTheme(cfg Config) Config {
	cfg.HomeFg = 255
	cfg.HomeBg = 105
	cfg.RodirFg = 255
	cfg.RodirBg = 124
	cfg.CwdOn = true
	cfg.CwdFg = 255
	cfg.CwdBg = 99
	cfg.RepoOn = true
	cfg.RepoFg = 0
	cfg.RepoBg = 255
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 255
	cfg.StatusSuccess = 41
	cfg.StatusError = 162
	cfg.StatusMisuse = 3
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 14
	cfg.StatusErrSignal = 202
	cfg.StatusTerminated = 240
	return cfg
}
