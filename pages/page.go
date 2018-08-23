package pages

import "math"

type Page struct {
	Total  int
	Offset int
	Index  int
	Size   int
	Length int
}

func NewPage(total int, size int, index int) *Page {
	page := &Page{
		Total: total,
		Size:  size,
		Index: index,
	}
	if page.Total < 0 {
		page.Total = 0
	}
	if page.Index < 1 {
		page.Index = 1
	}
	if page.Size < 1 {
		page.Size = 10
	}

	page.Offset = (page.Index - 1) * page.Size
	page.Length = int(math.Ceil(float64(page.Total) / float64(page.Size)))
	return page
}
