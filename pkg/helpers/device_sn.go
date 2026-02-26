package helpers

import (
	"fmt"
	"time"
)

// 34进制字符集 (0-9, A-Z 剔除 I, O)
const base34Chars = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// decimalToBase34 将10进制数转换为34进制（4位）
func decimalToBase34(num int) string {
	if num == 0 {
		return "0000"
	}

	digits := []byte{}
	n := num
	for n > 0 {
		remainder := n % 34
		digits = append([]byte{base34Chars[remainder]}, digits...)
		n /= 34
	}

	// 补足到4位
	for len(digits) < 4 {
		digits = append([]byte{'0'}, digits...)
	}

	return string(digits)
}

// getWeekNumber 获取当前年份和周数
func getWeekNumber(t time.Time) (year int, week int) {
	year, week = t.ISOWeek()
	return year % 100, week
}

// GenerateSN 生成JDP格式SN号
// 参数说明:
// productLine: 产品线代码(3位, 如001,002)
// pnCode: PN代码(6位)
// oemCode: OEM代码(2位, M1~MZ)
// serialNum: 流水号(十进制)
func GenerateSN(productLine string, pnCode string, oemCode string, serialNum int) string {
	// 获取当前生产时间
	now := time.Now()
	year, week := getWeekNumber(now)

	// 格式化生产时间: 年份后2位 + 周数(2位，补零)
	productionTime := fmt.Sprintf("%02d%02d", year, week)

	// 生成34进制流水号(4位)
	serialBase34 := decimalToBase34(serialNum)

	// 组合SN: JDP + 生产时间(4位) + PN代码(6位) + 产品线代码(3位) + OEM代码(2位) + 保留位ZZ(2位) + 流水号(4位)
	sn := fmt.Sprintf("JDP%s%s%s%sZZ%s",
		productionTime,
		pnCode,
		productLine,
		oemCode,
		serialBase34,
	)

	return sn
}
