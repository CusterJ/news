package server

import (
	"math"
	"os"
	"strconv"
)

func Pagination(page, total int) (pagelist []string) {
	var take, allPages, skip int
	take, _ = strconv.Atoi(os.Getenv("TAKE"))
	allPages = int(math.Ceil(float64(total) / float64(take)))
	// fmt.Printf("total/ take type is %T \n %v / %v = %v \n", total/take, total, take, total/take)
	if page > 0 {
		skip = (page - 1) * take
	}
	if skip > total {
		skip = 0
	}
	for i := 1; i <= allPages; i++ {
		pagelist = append(pagelist, strconv.Itoa(i))
	}
	// fmt.Printf("total= %v, take = %v, page = %v, allPages = %v, skip = %v \n", total, take, page, allPages, skip)
	return
}
