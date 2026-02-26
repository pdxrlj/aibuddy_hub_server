package helpers

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateSN(t *testing.T) {
	tests := []struct {
		name         string
		productLine  string
		pnCode       string
		oemCode      string
		serialNum    int
		wantLen      int    // 期望的SN长度
		wantContains string // 期望包含的字符串
	}{
		{
			name:         "serial 0",
			productLine:  "001",
			pnCode:       "123456",
			oemCode:      "M1",
			serialNum:    0,
			wantLen:      24,
			wantContains: "ZZ0000",
		},
		{
			name:         "serial 1",
			productLine:  "001",
			pnCode:       "123456",
			oemCode:      "M1",
			serialNum:    1,
			wantLen:      24,
			wantContains: "ZZ0001",
		},
		{
			name:         "serial 33",
			productLine:  "001",
			pnCode:       "123456",
			oemCode:      "M1",
			serialNum:    33,
			wantLen:      24,
			wantContains: "ZZ000Z",
		},
		{
			name:         "serial 34",
			productLine:  "001",
			pnCode:       "123456",
			oemCode:      "M1",
			serialNum:    34,
			wantLen:      24,
			wantContains: "ZZ0010",
		},
		{
			name:         "different product line and OEM",
			productLine:  "002",
			pnCode:       "789012",
			oemCode:      "M2",
			serialNum:    100,
			wantLen:      24,
			wantContains: "ZZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSN(tt.productLine, tt.pnCode, tt.oemCode, tt.serialNum)
			if len(got) != tt.wantLen {
				t.Errorf("GenerateSN() length = %v, want %v", len(got), tt.wantLen)
			}
			if !contains(got, tt.wantContains) {
				t.Errorf("GenerateSN() = %v, want to contain %v", got, tt.wantContains)
			}
			fmt.Printf("SN: %s\n", got)
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDecimalToBase34(t *testing.T) {
	// base34Chars = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ" (共34个字符, 去掉 I, O)
	// 位置: 0-9, A-H(10-17), J-N(18-23), P-T(24-27), U-W(28-30), X-Z(31-33)
	tests := []struct {
		name string
		num  int
		want string
	}{
		{"zero", 0, "0000"},
		{"one", 1, "0001"},
		{"ten", 10, "000A"},
		{"31", 31, "000X"},   // 位置 31 = X
		{"32", 32, "000Y"},   // 位置 32 = Y
		{"33", 33, "000Z"},   // 位置 33 = Z (最大值)
		{"34", 34, "0010"},   // 34 = 1*34 + 0
		{"100", 100, "002Y"}, // 100 = 2*34 + 32 = 68 + 32 = 100 (32=Y)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decimalToBase34(tt.num); got != tt.want {
				t.Errorf("decimalToBase34(%d) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}

func TestGetWeekNumber(t *testing.T) {
	// 使用固定时间测试
	testTime := time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC)
	year, week := getWeekNumber(testTime)
	// 2024年第50周
	if year != 24 {
		t.Errorf("year = %v, want 24", year)
	}
	if week != 50 {
		t.Errorf("week = %v, want 50", week)
	}
}
