package models_test

import (
	"testing"

	"fst/backend/app/models"
)

// TestSystemSetting_GetTypedValue_Number 回归：旧实现用 json.Marshal(字符串) 做无意义的判断，
// 且忽略 Unmarshal 错误，非法数字值会静默变成 0。改成 strconv.ParseFloat 后行为应清晰可预期。
func TestSystemSetting_GetTypedValue_Number(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  float64
	}{
		{"整数", "123", 123},
		{"小数", "3.14", 3.14},
		{"负数", "-5", -5},
		{"带空格", "  42  ", 42},
		{"非法值兜底为0", "not_a_number", 0},
		{"空字符串兜底为0", "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &models.SystemSetting{Type: "number", Value: c.value}
			got := s.GetTypedValue()
			gotFloat, ok := got.(float64)
			if !ok {
				t.Fatalf("GetTypedValue() 返回类型应为 float64，got %T", got)
			}
			if gotFloat != c.want {
				t.Errorf("GetTypedValue() = %v, want %v", gotFloat, c.want)
			}
		})
	}
}

func TestSystemSetting_GetTypedValue_Boolean(t *testing.T) {
	if v := (&models.SystemSetting{Type: "boolean", Value: "true"}).GetTypedValue(); v != true {
		t.Errorf("got %v, want true", v)
	}
	if v := (&models.SystemSetting{Type: "boolean", Value: "0"}).GetTypedValue(); v != false {
		t.Errorf("got %v, want false", v)
	}
}
