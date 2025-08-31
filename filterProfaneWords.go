package main

import "regexp"

func filterProfaneWords(str string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	res := str

	for _, word := range profaneWords {
		// Create case-insensitive pattern with word boundaries
		pattern := `(?i)\b` + regexp.QuoteMeta(word) + `\b`
		re := regexp.MustCompile(pattern)
		res = re.ReplaceAllString(res, "****")
	}
	return res
}
