package zptr

func StringToPtr(s string) *string {
	return &s
}

func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func BoolToPtr(b bool) *bool {
	return &b
}

func BoolOrDefault(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func IntToPtr(i int) *int {
	return &i
}

func IntOrDefault(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func Float32ToPtr(f float32) *float32 {
	return &f
}

func Float32OrDefault(f *float32) float32 {
	if f == nil {
		return 0.0
	}
	return *f
}

func Float64ToPtr(f float64) *float64 {
	return &f
}

func Float64OrDefault(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}
