package models

type LangContainer interface {
	CopyCodeToFile(*Container, Language, string) (string, error)
}
