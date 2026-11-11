package config

// ShowVersion 报告调用方是否只想查看版本。Load 在这种情况下返回的配置不能用于启动。
func ShowVersion(args []string) bool {
	return VersionRequested(args)
}
