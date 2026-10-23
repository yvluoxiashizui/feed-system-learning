package services

import (
	"testing"

	"feed/models"
)

func videoIDs(vs []models.Video) []uint {
	ids := make([]uint, 0, len(vs))
	for _, v := range vs {
		ids = append(ids, v.ID)
	}
	return ids
}

func equalIDs(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMergeByIDDesc(t *testing.T) {
	cases := []struct {
		name string
		a    []uint
		b    []uint
		lim  int
		want []uint
	}{
		{"两路都空", nil, nil, 5, nil},
		{"只有推来的", []uint{9, 7, 5}, nil, 5, []uint{9, 7, 5}},
		{"只有拉来的", nil, []uint{8, 6}, 5, []uint{8, 6}},
		{"交叉归并", []uint{9, 7, 5}, []uint{8, 6}, 5, []uint{9, 8, 7, 6, 5}},
		{"取前 limit 条", []uint{9, 7, 5}, []uint{8, 6}, 3, []uint{9, 8, 7}},
		{"同 id 去重", []uint{9, 7}, []uint{9, 5}, 5, []uint{9, 7, 5}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := make([]models.Video, 0, len(c.a))
			for _, id := range c.a {
				a = append(a, models.Video{ID: id})
			}
			b := make([]models.Video, 0, len(c.b))
			for _, id := range c.b {
				b = append(b, models.Video{ID: id})
			}

			got := videoIDs(mergeByIDDesc(a, b, c.lim))
			if len(c.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("期望空结果，实际 %v", got)
				}
				return
			}
			if !equalIDs(got, c.want) {
				t.Fatalf("结果 %v，期望 %v", got, c.want)
			}
		})
	}
}
