package search

import (
	"context"
	"fmt"
	"kcdb/db"
	"os"
	"sort"
	"strconv"
	"strings"
)

type byRankSym []*db.Symbol

func (a byRankSym) Len() int           { return len(a) }
func (a byRankSym) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRankSym) Less(i, j int) bool { return a[i].Rank < a[j].Rank }

func rankSym(ctx context.Context, syms []*db.Symbol) ([]*db.Symbol, error) {
	var s byRankSym
	for _, fp := range syms {
		src, err := getSource(ctx, fp.SourceID)
		if err != nil {
			if err == os.ErrNotExist {
				//TODO: We should fix the data inconsistency instead.
				src = &db.Source{Rank: -200}
			} else {
				return nil, err
			}
		}
		fp.Rank = -src.Rank
	}
	s = byRankSym(syms)
	sort.Sort(s)
	return s, nil
}

// SymbolSearch returns search results when searching for symbols.
func SymbolSearch(ctx context.Context, q string) ([]*db.Symbol, error) {
	var params db.SymSearchParam
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
			default:
				return nil, fmt.Errorf("could not understand specifier %q", spl[0])
			}
		} else {
			params.Keywords = append(params.Keywords, token)
		}
	}

	if len(params.Keywords) == 0 {
		return nil, ErrBadQuery{msg: "Keywords must be specified"}
	}

	syms, err := db.SymbolSearch(ctx, params, db.DB())
	if err != nil {
		return nil, err
	}
	return rankSym(ctx, syms)
}
