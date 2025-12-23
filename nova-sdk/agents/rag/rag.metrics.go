package rag

import (
	"time"
)

// Metrics tracks performance and usage metrics for RAG embedding operations
type Metrics struct {
	TotalEmbeddings  int
	TotalDimensions  int
	TotalProcessTime time.Duration
	TotalCharacters  int
	SearchOperations int
	TotalSearchTime  time.Duration
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordEmbedding records metrics for a single embedding operation
func (m *Metrics) RecordEmbedding(content string, dimensions int, duration time.Duration) {
	m.TotalEmbeddings++
	m.TotalDimensions += dimensions
	m.TotalProcessTime += duration
	m.TotalCharacters += len(content)
}

// RecordSearch records metrics for a search operation
func (m *Metrics) RecordSearch(duration time.Duration) {
	m.SearchOperations++
	m.TotalSearchTime += duration
}

// AvgDimensions returns the average number of dimensions per embedding
func (m *Metrics) AvgDimensions() int {
	if m.TotalEmbeddings == 0 {
		return 0
	}
	return m.TotalDimensions / m.TotalEmbeddings
}

// AvgEmbeddingTime returns the average time per embedding operation
func (m *Metrics) AvgEmbeddingTime() time.Duration {
	if m.TotalEmbeddings == 0 {
		return 0
	}
	return m.TotalProcessTime / time.Duration(m.TotalEmbeddings)
}

// AvgCharsPerDocument returns the average number of characters per document
func (m *Metrics) AvgCharsPerDocument() int {
	if m.TotalEmbeddings == 0 {
		return 0
	}
	return m.TotalCharacters / m.TotalEmbeddings
}

// AvgSearchTime returns the average time per search operation
func (m *Metrics) AvgSearchTime() time.Duration {
	if m.SearchOperations == 0 {
		return 0
	}
	return m.TotalSearchTime / time.Duration(m.SearchOperations)
}

// TotalOperations returns the total number of operations (embeddings + searches)
func (m *Metrics) TotalOperations() int {
	return m.TotalEmbeddings + m.SearchOperations
}

// TotalTime returns the total time spent on all operations
func (m *Metrics) TotalTime() time.Duration {
	return m.TotalProcessTime + m.TotalSearchTime
}

// AvgOperationTime returns the average time per operation (embedding or search)
func (m *Metrics) AvgOperationTime() time.Duration {
	totalOps := m.TotalOperations()
	if totalOps == 0 {
		return 0
	}
	return m.TotalTime() / time.Duration(totalOps)
}

// Throughput returns operations per second
func (m *Metrics) Throughput() float64 {
	totalTime := m.TotalTime()
	if totalTime.Seconds() == 0 {
		return 0
	}
	return float64(m.TotalOperations()) / totalTime.Seconds()
}

// EstimateCost estimates the cost of embeddings based on character count
// costPerThousandChars is the cost per 1000 characters (e.g., 0.0001 USD)
func (m *Metrics) EstimateCost(costPerThousandChars float64) float64 {
	return float64(m.TotalCharacters) / 1000.0 * costPerThousandChars
}

// CostPerDocument returns the estimated cost per document
// costPerThousandChars is the cost per 1000 characters (e.g., 0.0001 USD)
func (m *Metrics) CostPerDocument(costPerThousandChars float64) float64 {
	if m.TotalEmbeddings == 0 {
		return 0
	}
	return m.EstimateCost(costPerThousandChars) / float64(m.TotalEmbeddings)
}

// Reset clears all metrics
func (m *Metrics) Reset() {
	m.TotalEmbeddings = 0
	m.TotalDimensions = 0
	m.TotalProcessTime = 0
	m.TotalCharacters = 0
	m.SearchOperations = 0
	m.TotalSearchTime = 0
}
