package models

type LanguageDockerImageName string

const (
	CppImage    LanguageDockerImageName = "gcc:latest"
	PythonImage LanguageDockerImageName = "python:3.11-alpine"
)
