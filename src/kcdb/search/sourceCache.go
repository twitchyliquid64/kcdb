package search

import (
  "context"
  "time"

  "kcdb/db"
)

var cachedSources = map[int]*db.Source{}

func getSource(ctx context.Context, sourceUid int) (*db.Source, error) {
  cv, ok := cachedSources[sourceUid]
  if (time.Now().Unix() % 20 == 0) || !ok {
    v, err := db.GetSource(ctx, sourceUid, db.DB())
    if err != nil {
      return nil, err
    }
    cachedSources[sourceUid] = v
    return v, nil
  }
  return cv, nil
}
