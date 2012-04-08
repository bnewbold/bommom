package main

type EmailAddress string
type Password string
type Url string

// "Slug" string with limited ASCII character set, good for URLs.
// Lowercase alphanumeric plus '_' allowed. 
type ShortName string
