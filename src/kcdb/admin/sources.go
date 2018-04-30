package admin

import (
  "net/http"
  "strconv"
  "fmt"

  "kcdb/db"
)

// UpdateSourceAdmin is called to set the tag & rank of a source.
func UpdateSourceAdmin(w http.ResponseWriter, req *http.Request) {
  if adminSecret == "" {
    return
  }
  if req.FormValue("secret") != adminSecret {
    http.Error(w, "Not Authorized", http.StatusUnauthorized)
    return
  }
  rank, err := strconv.Atoi(req.FormValue("rank"))
  if err != nil {
    http.Error(w, "Bad request", http.StatusBadRequest)
    fmt.Printf("Err: %v\n", err)
    return
  }
  uid, err := strconv.Atoi(req.FormValue("uid"))
  if err != nil {
    http.Error(w, "Bad request", http.StatusBadRequest)
    fmt.Printf("Err: %v\n", err)
    return
  }
  err = db.SetSourceAdmin(req.Context(), uid, rank, req.FormValue("tag"), db.DB())
  if err != nil {
    http.Error(w, "Internal error", http.StatusInternalServerError)
    fmt.Printf("Err: %v\n", err)
    return
  }
  w.Write([]byte("OK."))
}

// AddSourceAdmin is called to add a source.
func AddSourceAdmin(w http.ResponseWriter, req *http.Request) {
  if adminSecret == "" {
    return
  }
  if req.FormValue("secret") != adminSecret {
    http.Error(w, "Not Authorized", http.StatusUnauthorized)
    return
  }
  err := db.AddSource(req.Context(), &db.Source{Kind: "git", URL: req.FormValue("tag")}, db.DB())
  if err != nil {
    http.Error(w, "Internal error", http.StatusInternalServerError)
    fmt.Printf("Err: %v\n", err)
    return
  }
  w.Write([]byte("OK."))
}
