package orm

const (
	asc  = "ASC"
	desc = "DESC"
)

type OrderBy struct {
	col   string
	order string
}

func Asc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: asc,
	}
}

func Desc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: desc,
	}
}
