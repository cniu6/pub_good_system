package admin

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestDecodeTerminalOutput_UTF8Passthrough(t *testing.T) {
	in := []byte("驱动器卷的序列号是 1234")
	got := decodeTerminalOutput(in)
	if got != string(in) {
		t.Fatalf("utf8 passthrough: got %q", got)
	}
}

func TestDecodeTerminalOutput_GBKChinese(t *testing.T) {
	// 「驱动器」的 GBK 字节，模拟中文 Windows cmd 输出
	gbk, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte("驱动器 C 中的卷是 Windows"))
	if err != nil {
		t.Fatal(err)
	}
	got := decodeTerminalOutput(gbk)
	if got != "驱动器 C 中的卷是 Windows" {
		t.Fatalf("gbk decode: got %q", got)
	}
}
