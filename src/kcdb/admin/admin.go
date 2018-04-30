package admin

var adminSecret = ""

// SetSecret sets the password for admin RPCs.
func SetSecret(s string) {
  adminSecret = s
}
