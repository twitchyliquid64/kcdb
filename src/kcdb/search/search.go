package search

import (
	"context"
	"fmt"
	"kcdb/db"
	"strconv"
	"strings"
)

// Search returns search results.
func Search(ctx context.Context, q string) ([]*db.Footprint, error) {
	var params db.FpSearchParam
	var err error

	for _, token := range strings.Split(q, " ") {
		if strings.Contains(token, "=") {
			spl := strings.Split(token, "=")
			switch spl[0] {
			case "pin_count":
				params.PinCount, err = strconv.Atoi(spl[1])
				if err != nil {
					return nil, err
				}
			case "attr":
				params.Attr = spl[1]
			default:
				return nil, fmt.Errorf("could not understand specifier %q", spl[0])
			}
		} else {
			params.Keywords = append(params.Keywords, token)
		}
	}

	return db.FootprintSearch(ctx, params, db.DB())
}
