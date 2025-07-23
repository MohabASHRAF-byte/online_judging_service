package models

type LangContainer interface {
	CopyCodeToFile(*Container, Language, string) (string, error)
	CompileCode(*Container, string) (string, error)
	RunTestCases(*Container, []string, string) ([]string, error)
}
