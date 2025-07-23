package models

import "context"

type LangContainer interface {
	CopyCodeToFile(*Container, string) (string, error)
	CompileCode(*Container, string, context.Context) (string, error)
	RunTestCases(*Container, string, string, context.Context) (string, error)
}
