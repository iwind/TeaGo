package stringutil

import "testing"

func TestRandString(t *testing.T) {
	t.Log(Rand(10))
}

func TestConvertID(t *testing.T) {
	t.Log(ConvertID(1234567890))
}

func TestVersionCompare(t *testing.T) {
	t.Log(VersionCompare("1.0", "1.0.3"))
	t.Log(VersionCompare("2.0.3", "2.0.3"))
	t.Log(VersionCompare("2", "2.1"))
	t.Log(VersionCompare("1.1.2", "1.2.1"))
	t.Log(VersionCompare("1.10.2", "1.2.1"))
	t.Log(VersionCompare("1.14.2", "1.1234567.1"))
}