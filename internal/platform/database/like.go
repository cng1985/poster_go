package database

import "strings"

// Contains 生成转义后的包含匹配。列名必须来自白名单，不能来自客户端。
func Contains(column, value string) (string, string) {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
	return column + " like ? escape '\\'", "%" + escaped + "%"
}
