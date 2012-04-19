package main

// Minimal Error type... is there a better way?
type Error string

func (e Error) Error() string {
	return string(e)
}

type EmailAddress string
type Password string
type Url string

// "Slug" string with limited ASCII character set, good for URLs.
// Lowercase alphanumeric plus '_' allowed. 
type ShortName string

func isShortName(s string) bool {
	for i, r := range s {
		switch {
		case '0' <= r && '9' >= r && i > 0:
			continue
		case 'a' <= r && 'z' >= r:
			continue
		case r == '_' && i > 0:
			continue
		default:
			return false
		}
	}
	return len(s) > 0
}
