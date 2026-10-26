package jev

import (
	"context"
	"sync"
)

var (
	defaultClassifier Classifier
	initOnce          sync.Once
)

// Default returns the process-wide default JEV classifier (Runner with Local fallback).
func Default() Classifier {
	initOnce.Do(func() {
		local := NewLocalClassifier()
		runner := NewRunnerClassifier(local)
		if runner.IsAvailable() {
			defaultClassifier = runner
		} else {
			defaultClassifier = local
		}
	})
	return defaultClassifier
}

// SetDefault allows overriding the default classifier (e.g. for testing).
func SetDefault(c Classifier) {
	defaultClassifier = c
}

// ClassifyIntent routes a natural language query using the default JEV classifier.
func ClassifyIntent(query string) IntentResult {
	return Default().ClassifyIntent(context.Background(), query)
}

// ClassifyDocument categorizes a markdown or spec file using the default JEV classifier.
func ClassifyDocument(path string, content []byte) DocumentResult {
	return Default().ClassifyDocument(context.Background(), path, content)
}

// ClassifyDrift analyzes the severity of divergence between specification and code.
func ClassifyDrift(specSummary string, diffSummary string) DriftResult {
	return Default().ClassifyDrift(context.Background(), specSummary, diffSummary)
}
