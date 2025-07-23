package models

type LangContainer interface {
	CopyCodeToFile(*Container, string) (string, error)
	CompileCode(*Container, string) (string, error)
	RunTestCases(*Container, []string, string) ([]string, error)
}
