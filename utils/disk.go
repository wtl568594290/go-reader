package utils

import "os"

func HasDisk(disk string) bool {
	// 判断 磁盘是否存在
	_, err := os.Stat(disk)
	if err != nil {
		return false
	}
	// 如果存在，且不为空，则返回true
	if _, err := os.ReadDir(disk); err == nil {
		return true
	}
	return false
}
