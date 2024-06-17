package env

import "os"

func init() {
	// 设置环境变量，否则ligloss会出现中文字符宽度不一致的问题
	os.Setenv("RUNEWIDTH_EASTASIAN", "0")
}
