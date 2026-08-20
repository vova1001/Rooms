package internal

import "net/http"

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true, // true при HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}
