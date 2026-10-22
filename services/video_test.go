package services

import "testing"

// 窗口按 id 降序排列，和线上缓存里的顺序一致
var windowIDs = []uint{100, 99, 98, 97, 96, 95, 94, 93, 92, 91}

func TestWindowSlice(t *testing.T) {
	cases := []struct {
		name    string
		ids     []uint
		cursor  uint
		limit   int
		wantIDs []uint
		wantOK  bool
	}{
		{
			name:    "首页从窗口头部开始",
			ids:     windowIDs,
			cursor:  0,
			limit:   3,
			wantIDs: []uint{100, 99, 98},
			wantOK:  true,
		},
		{
			name:    "游标在窗口内部",
			ids:     windowIDs,
			cursor:  96,
			limit:   3,
			wantIDs: []uint{95, 94, 93},
			wantOK:  true,
		},
		{
			name:    "刚好取到窗口末尾",
			ids:     windowIDs,
			cursor:  93,
			limit:   2,
			wantIDs: []uint{92, 91},
			wantOK:  true,
		},
		{
			name:   "没凑满一页且到窗口末尾，判定为可能被截断",
			ids:    windowIDs,
			cursor: 95,
			limit:  10,
			wantOK: false,
		},
		{
			name:   "游标比窗口最老的 id 还小",
			ids:    windowIDs,
			cursor: 50,
			limit:  3,
			wantOK: false,
		},
		{
			name:   "空窗口",
			ids:    nil,
			cursor: 0,
			limit:  10,
			wantOK: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := windowSlice(c.ids, c.cursor, c.limit)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v", ok, c.wantOK)
			}
			if !ok {
				return
			}
			if len(got) != len(c.wantIDs) {
				t.Fatalf("返回 %d 条，期望 %d 条: %v", len(got), len(c.wantIDs), got)
			}
			for i := range c.wantIDs {
				if got[i] != c.wantIDs[i] {
					t.Fatalf("第 %d 个 id = %d，期望 %d", i, got[i], c.wantIDs[i])
				}
			}
		})
	}
}
