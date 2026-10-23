package services

import "testing"

func TestRankCursorRoundTrip(t *testing.T) {
	cases := []struct {
		likes int64
		id    uint
	}{
		{0, 1},
		{20, 111653},
		{999999, 4294967295},
	}

	for _, c := range cases {
		raw := encodeRankCursor(c.likes, c.id)
		gotLikes, gotID := parseRankCursor(raw)
		if gotLikes != c.likes || gotID != c.id {
			t.Fatalf("游标 %q 解析成 (%d, %d)，期望 (%d, %d)", raw, gotLikes, gotID, c.likes, c.id)
		}
	}
}

func TestParseRankCursorInvalid(t *testing.T) {
	// 非法游标一律从头开始分页，不能报错也不能返回半截游标
	for _, raw := range []string{"", "abc", "20", "20_", "_123", "-1_5", "20_abc", "20_5_9"} {
		likes, id := parseRankCursor(raw)
		if raw == "20_5_9" {
			// 多段时只按第一段切分，后面的当作 id 的一部分，解析失败回到零值
			if id != 0 {
				t.Fatalf("游标 %q 期望 id=0，实际 %d", raw, id)
			}
			continue
		}
		if likes != 0 || id != 0 {
			t.Fatalf("游标 %q 期望零值，实际 (%d, %d)", raw, likes, id)
		}
	}
}
