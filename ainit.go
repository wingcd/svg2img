package main

var (
	CONFIG *IniParser
)

func init() {
	CONFIG = new(IniParser)

	CONFIG.Load("./config.ini")

	DefaultChromPaths = append(DefaultChromPaths, CONFIG.GetString("", "chrome"))
}
