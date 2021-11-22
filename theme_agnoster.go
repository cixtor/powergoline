package main

func ApplyAgnosterTheme(cfg Config) Config {
	cfg.UserOn = true
	cfg.UserFg = 255
	cfg.UserBg = 26
	cfg.HomeFg = 255
	cfg.HomeBg = 161
	cfg.RodirFg = 255
	cfg.RodirBg = 124
	cfg.CwdOn = true
	cfg.CwdN = 2
	cfg.CwdFg = 8
	cfg.CwdBg = 255
	cfg.SymbolUser = "$"
	cfg.SymbolRoot = "#"
	cfg.StatusFg = 255
	cfg.StatusSuccess = 72
	cfg.StatusError = 88
	cfg.StatusMisuse = 94
	cfg.StatusCantExec = 4
	cfg.StatusNotFound = 38
	cfg.StatusErrSignal = 130
	cfg.StatusTerminated = 13
	return cfg
}
