package data

type Filter struct {
	IDs    []int
	Offset int
	Limit  int
}

func NewFilter(offset, limit int, ids []int) Filter {
	return Filter{
		IDs:    ids,
		Offset: offset,
		Limit:  limit,
	}
}
