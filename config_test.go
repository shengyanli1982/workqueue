package workqueue

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// legacyRetryKey 是 P1-1 类型特化优化前的默认 keyFunc 实现，
// 作为等价性断言的金标准：特化快路径的输出必须与其逐字节一致。
func legacyRetryKey(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%T:%#v", value, value)
}

func assertKeyEquivalence[T any](t *testing.T, cases []T) {
	t.Helper()
	for _, c := range cases {
		assert.Equal(t, legacyRetryKey(c), defaultRetryKeyFunc(c),
			"keyFunc output must be byte-identical to fmt.Sprintf(\"%%T:%%#v\")")
	}
}

func TestDefaultRetryKeyFunc_Equivalence_String(t *testing.T) {
	cases := []string{
		"", "a", "task", "hello world",
		`"quoted"`, `back\slash`, "new\nline", "tab\there", "cr\rreturn\r\n",
		"中文", "日本語テキスト", "한국어", "🚀🌟", "mixed混合ascii123!@#",
		"\x00", "\x7f", "\x1b[31mesc\x1b[0m",
		"\xff\xfe", "invalid\x80utf8", "\xed\xa0\x80surrogate",
		"'single'", "%v %#v %T %s", "{braces}", "[brackets]", "key:colon",
		strings.Repeat("x", 200), "\u00e9\u0301combining", " trailing ",
	}
	assertKeyEquivalence(t, cases)
}

func TestDefaultRetryKeyFunc_Equivalence_Int(t *testing.T) {
	cases := []int{
		0, 1, -1, 2, -2, 9, 10, -10, 42, -42, 99, 100, 255, 256, -256,
		123456789, -123456789, 1000000007, -1000000007,
		math.MaxInt32, math.MinInt32, math.MaxInt32 + 1, math.MinInt32 - 1,
		math.MaxInt, math.MinInt, math.MaxInt - 1, math.MinInt + 1,
	}
	assertKeyEquivalence(t, cases)
}

func TestDefaultRetryKeyFunc_Equivalence_Int64(t *testing.T) {
	cases := []int64{
		0, 1, -1, 42, -42, 100, -100, 255, 256,
		math.MaxInt32, math.MinInt32, math.MaxInt32 + 1, math.MinInt32 - 1,
		1 << 32, -(1 << 32), 1 << 62, -(1 << 62),
		999999999999999999, -999999999999999999,
		123456789012345, -123456789012345,
		math.MaxInt64, math.MinInt64, math.MaxInt64 - 1, math.MinInt64 + 1,
	}
	assertKeyEquivalence(t, cases)
}

func TestDefaultRetryKeyFunc_Equivalence_Uint64(t *testing.T) {
	cases := []uint64{
		0, 1, 2, 42, 255, 256, 36, 35, 37, 1295, 1296,
		math.MaxUint32, math.MaxUint32 + 1, 1 << 32, 1 << 63, 1<<63 - 1,
		uint64(math.MaxInt64), uint64(math.MaxInt64) + 1,
		999999999999999999, 123456789012345, 1000000007,
		math.MaxUint64 - 1, math.MaxUint64,
	}
	assertKeyEquivalence(t, cases)
}

func TestDefaultRetryKeyFunc_Equivalence_Bool(t *testing.T) {
	assertKeyEquivalence(t, []bool{true, false})
}

// eqFallbackStruct 验证非特化类型仍走 Sprintf 回退路径且输出不变。
type eqFallbackStruct struct {
	name string
	age  int
}

func TestDefaultRetryKeyFunc_Equivalence_FallbackTypes(t *testing.T) {
	assert.Equal(t, "", defaultRetryKeyFunc(nil), "nil must produce empty key")
	assert.Equal(t, legacyRetryKey(nil), defaultRetryKeyFunc(nil))

	assertKeyEquivalence(t, []any{
		3.14, float32(1.5), int32(7), uint(9), int8(-3), uint8(255), int16(-300), uint16(65535),
		eqFallbackStruct{name: "x", age: 1}, &eqFallbackStruct{},
		[]int{1, 2, 3}, map[string]int{"k": 1}, [2]byte{1, 2},
	})
}

// TestDefaultRetryKeyFunc_Equivalence_Random 属性测试：每类型 1000 个随机值，
// 特化路径输出必须与 fmt.Sprintf("%T:%#v") 逐字节一致。
func TestDefaultRetryKeyFunc_Equivalence_Random(t *testing.T) {
	rng := rand.New(rand.NewSource(20260912))

	for i := 0; i < 1000; i++ {
		// 随机字符串：全长域字节（含非法 UTF-8 序列）。
		n := rng.Intn(33)
		buf := make([]byte, n)
		for j := range buf {
			buf[j] = byte(rng.Intn(256))
		}
		s := string(buf)
		assert.Equal(t, legacyRetryKey(s), defaultRetryKeyFunc(s), "random string case")

		v64 := rng.Uint64()
		assert.Equal(t, legacyRetryKey(v64), defaultRetryKeyFunc(v64), "random uint64 case")

		i64 := rng.Int63()
		if rng.Intn(2) == 0 {
			i64 = -i64
		}
		assert.Equal(t, legacyRetryKey(i64), defaultRetryKeyFunc(i64), "random int64 case")

		iv := int(i64)
		assert.Equal(t, legacyRetryKey(iv), defaultRetryKeyFunc(iv), "random int case")

		b := rng.Intn(2) == 0
		assert.Equal(t, legacyRetryKey(b), defaultRetryKeyFunc(b), "random bool case")
	}
}
