package server

import (
	"math"
	"os"
	"strconv"
)

func Pagination(page, total int) (pagelist []string) {
	var take, allPages int
	take, err := strconv.Atoi(os.Getenv("TAKE"))
	if err != nil {
		take = 15 // set default
	}
	allPages = int(math.Ceil(float64(total) / float64(take)))

	if page-3 > 2 {
		pagelist = append(pagelist, strconv.Itoa(1))
	}

	for i := 1; i <= allPages; i++ {
		if i >= page-3 && i <= page+3 {
			pagelist = append(pagelist, strconv.Itoa(i))
			continue
		}
		if i >= page-30 && i <= page+30 && i%10 == 0 {
			pagelist = append(pagelist, strconv.Itoa(i))
		}
	}

	if page+3 < allPages && allPages%10 != 0 {
		pagelist = append(pagelist, strconv.Itoa(allPages))
	}

	return
}
