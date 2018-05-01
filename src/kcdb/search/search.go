package search

import (
	"context"
	"fmt"
	"kcdb/db"
	"sort"
	"strconv"
	"strings"
)

type byRank []*db.Footprint

func (a byRank) Len() int           { return len(a) }
func (a byRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRank) Less(i, j int) bool { return a[i].Rank < a[j].Rank }

func rank(ctx context.Context, fps []*db.Footprint) ([]*db.Footprint, error) {
	var s byRank
	for _, fp := range fps {
		src, err := getSource(ctx, fp.SourceID)
		if err != nil {
			return nil, err
		}
		fp.Rank = src.Rank
	}
	s = byRank(fps)
	sort.Sort(s)
	return s, nil
}

// Search returns search results.
func Search(ctx context.Context, q string) ([]*db.Footprint, error) {
	var params db.FpSearchParam
	var err error

	for _, token := range strings.Split(q, " ") {
		if strings.Contains(token, "=") {
			spl := strings.Split(token, "=")
			switch spl[0] {
			case "pin_count", "pc", "pinc", "pin_c", "p_count", "pin_cnt":
				params.PinCount, err = strconv.Atoi(spl[1])
				if err != nil {
					return nil, err
				}
			case "attr", "at", "attribute":
				params.Attr = spl[1]
			default:
				return nil, fmt.Errorf("could not understand specifier %q", spl[0])
			}
		} else {
			params.Keywords = append(params.Keywords, token)
		}
	}

	fps, err := db.FootprintSearch(ctx, params, db.DB())
	if err != nil {
		return nil, err
	}
	return rank(ctx, fps)
}
