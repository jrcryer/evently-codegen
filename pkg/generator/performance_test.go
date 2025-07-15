package generator

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkParseSmallSchema benchmarks parsing of small AsyncAPI schemas
func BenchmarkParseSmallSchema(b *testing.B) {
	spec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Small Schema", "version": "1.0.0"},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"name": {"type": "string"},
						"email": {"type": "string"}
					}
				}
			}
		}
	}`

	config := &Config{
		PackageName: "benchmark",
		OutputDir:   "./test",
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Parse([]byte(spec))
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkParseMediumSchema benchmarks parsing of medium-sized AsyncAPI schemas
func BenchmarkParseMediumSchema(b *testing.B) {
	// Generate a medium-sized schema with 20 properties
	properties := make(map[string]interface{})
	for i := 0; i < 20; i++ {
		properties[fmt.Sprintf("field%d", i)] = map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("Field %d description", i),
		}
	}

	spec := map[string]interface{}{
		"asyncapi": "2.6.0",
		"info":     map[string]interface{}{"title": "Medium Schema", "version": "1.0.0"},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"MediumSchema": map[string]interface{}{
					"type":       "object",
					"properties": properties,
				},
			},
		},
	}

	specBytes, _ := json.Marshal(spec)
	config := &Config{
		PackageName: "benchmark",
		OutputDir:   "./test",
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Parse(specBytes)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkParseLargeSchema benchmarks parsing of large AsyncAPI schemas
func BenchmarkParseLargeSchema(b *testing.B) {
	// Generate a large schema with 100 properties
	properties := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		var propType string
		switch i % 4 {
		case 0:
			propType = "string"
		case 1:
			propType = "integer"
		case 2:
			propType = "number"
		case 3:
			propType = "boolean"
		}

		properties[fmt.Sprintf("field%d", i)] = map[string]interface{}{
			"type":        propType,
			"description": fmt.Sprintf("Field %d description with longer text to simulate real-world schemas", i),
		}
	}

	spec := map[string]interface{}{
		"asyncapi": "2.6.0",
		"info":     map[string]interface{}{"title": "Large Schema", "version": "1.0.0"},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"LargeSchema": map[string]interface{}{
					"type":       "object",
					"properties": properties,
				},
			},
		},
	}

	specBytes, _ := json.Marshal(spec)
	config := &Config{
		PackageName: "benchmark",
		OutputDir:   "./test",
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Parse(specBytes)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkGenerateSmallSchema benchmarks code generation for small schemas
func BenchmarkGenerateSmallSchema(b *testing.B) {
	messages := map[string]*MessageSchema{
		"User": {
			Name:        "User",
			Type:        "object",
			Description: "User information",
			Properties: map[string]*Property{
				"id":    {Type: "string", Description: "User ID"},
				"name":  {Type: "string", Description: "User name"},
				"email": {Type: "string", Description: "User email"},
			},
			Required: []string{"id"},
		},
	}

	config := &Config{
		PackageName:     "benchmark",
		OutputDir:       "./test",
		IncludeComments: true,
		UsePointers:     true,
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(messages)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

// BenchmarkGenerateLargeSchema benchmarks code generation for large schemas
func BenchmarkGenerateLargeSchema(b *testing.B) {
	properties := make(map[string]*Property)
	required := make([]string, 0)

	for i := 0; i < 100; i++ {
		propName := fmt.Sprintf("field%d", i)
		var propType string
		switch i % 4 {
		case 0:
			propType = "string"
		case 1:
			propType = "integer"
		case 2:
			propType = "number"
		case 3:
			propType = "boolean"
		}

		properties[propName] = &Property{
			Type:        propType,
			Description: fmt.Sprintf("Field %d description", i),
		}

		if i%10 == 0 {
			required = append(required, propName)
		}
	}

	messages := map[string]*MessageSchema{
		"LargeSchema": {
			Name:        "LargeSchema",
			Type:        "object",
			Description: "Large schema for benchmarking",
			Properties:  properties,
			Required:    required,
		},
	}

	config := &Config{
		PackageName:     "benchmark",
		OutputDir:       "./test",
		IncludeComments: true,
		UsePointers:     true,
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(messages)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

// BenchmarkParseAndGenerate benchmarks the complete workflow
func BenchmarkParseAndGenerate(b *testing.B) {
	spec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Benchmark API", "version": "1.0.0"},
		"components": {
			"schemas": {
				"Message": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"timestamp": {"type": "string", "format": "date-time"},
						"data": {"type": "string"},
						"metadata": {
							"type": "object",
							"properties": {
								"source": {"type": "string"},
								"version": {"type": "string"}
							}
						}
					},
					"required": ["id", "timestamp"]
				}
			}
		}
	}`

	config := &Config{
		PackageName:     "benchmark",
		OutputDir:       "./test",
		IncludeComments: true,
		UsePointers:     true,
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.ParseAndGenerate([]byte(spec))
		if err != nil {
			b.Fatalf("ParseAndGenerate failed: %v", err)
		}
	}
}

// BenchmarkMultipleSchemas benchmarks generation with multiple schemas
func BenchmarkMultipleSchemas(b *testing.B) {
	schemas := make(map[string]interface{})

	// Create 10 different schemas
	for i := 0; i < 10; i++ {
		properties := make(map[string]interface{})
		for j := 0; j < 10; j++ {
			properties[fmt.Sprintf("field%d", j)] = map[string]interface{}{
				"type": "string",
			}
		}

		schemas[fmt.Sprintf("Schema%d", i)] = map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}
	}

	spec := map[string]interface{}{
		"asyncapi": "2.6.0",
		"info":     map[string]interface{}{"title": "Multi Schema", "version": "1.0.0"},
		"components": map[string]interface{}{
			"schemas": schemas,
		},
	}

	specBytes, _ := json.Marshal(spec)
	config := &Config{
		PackageName: "benchmark",
		OutputDir:   "./test",
	}
	gen := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.ParseAndGenerate(specBytes)
		if err != nil {
			b.Fatalf("ParseAndGenerate failed: %v", err)
		}
	}
}

// TestMemoryUsage tests memory usage with large schemas
func TestMemoryUsage(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Generate a very large schema
	properties := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		properties[fmt.Sprintf("field%d", i)] = map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("Field %d with a long description to increase memory usage", i),
		}
	}

	spec := map[string]interface{}{
		"asyncapi": "2.6.0",
		"info":     map[string]interface{}{"title": "Memory Test", "version": "1.0.0"},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"LargeSchema": map[string]interface{}{
					"type":       "object",
					"properties": properties,
				},
			},
		},
	}

	specBytes, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal spec: %v", err)
	}

	config := &Config{
		PackageName:     "memorytest",
		OutputDir:       "./test",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := NewGenerator(config)

	// Parse and generate
	result, err := gen.ParseAndGenerate(specBytes)
	if err != nil {
		t.Fatalf("ParseAndGenerate failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("No files generated")
	}

	// Check memory usage after generation
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryUsed := m2.Alloc - m1.Alloc
	t.Logf("Memory used for large schema generation: %d bytes (%.2f MB)",
		memoryUsed, float64(memoryUsed)/(1024*1024))

	// Verify memory usage is reasonable (less than 50MB for this test)
	maxMemoryMB := float64(50)
	actualMemoryMB := float64(memoryUsed) / (1024 * 1024)
	if actualMemoryMB > maxMemoryMB {
		t.Errorf("Memory usage too high: %.2f MB (max: %.2f MB)", actualMemoryMB, maxMemoryMB)
	}
}

// TestProcessingSpeed tests processing speed with various schema sizes
func TestProcessingSpeed(t *testing.T) {
	testCases := []struct {
		name       string
		fieldCount int
		maxTime    time.Duration
	}{
		{"small_schema", 10, 10 * time.Millisecond},
		{"medium_schema", 50, 50 * time.Millisecond},
		{"large_schema", 200, 200 * time.Millisecond},
		{"very_large_schema", 500, 500 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate schema with specified field count
			properties := make(map[string]interface{})
			for i := 0; i < tc.fieldCount; i++ {
				properties[fmt.Sprintf("field%d", i)] = map[string]interface{}{
					"type":        "string",
					"description": fmt.Sprintf("Field %d description", i),
				}
			}

			spec := map[string]interface{}{
				"asyncapi": "2.6.0",
				"info":     map[string]interface{}{"title": "Speed Test", "version": "1.0.0"},
				"components": map[string]interface{}{
					"schemas": map[string]interface{}{
						"SpeedTestSchema": map[string]interface{}{
							"type":       "object",
							"properties": properties,
						},
					},
				},
			}

			specBytes, err := json.Marshal(spec)
			if err != nil {
				t.Fatalf("Failed to marshal spec: %v", err)
			}

			config := &Config{
				PackageName: "speedtest",
				OutputDir:   "./test",
			}
			gen := NewGenerator(config)

			// Measure processing time
			start := time.Now()
			result, err := gen.ParseAndGenerate(specBytes)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("ParseAndGenerate failed: %v", err)
			}

			if len(result.Files) == 0 {
				t.Fatal("No files generated")
			}

			t.Logf("Processing %d fields took %v", tc.fieldCount, duration)

			// Verify processing time is within acceptable limits
			if duration > tc.maxTime {
				t.Errorf("Processing took too long: %v (max: %v)", duration, tc.maxTime)
			}
		})
	}
}

// TestConcurrentGeneration tests concurrent generation safety
func TestConcurrentGeneration(t *testing.T) {
	spec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Concurrent Test", "version": "1.0.0"},
		"components": {
			"schemas": {
				"ConcurrentSchema": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"data": {"type": "string"}
					}
				}
			}
		}
	}`

	config := &Config{
		PackageName: "concurrent",
		OutputDir:   "./test",
	}

	// Run multiple generators concurrently
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			gen := NewGenerator(config)
			_, err := gen.ParseAndGenerate([]byte(spec))
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent generation %d failed: %v", i, err)
		}
	}
}

// TestMemoryLeaks tests for potential memory leaks
func TestMemoryLeaks(t *testing.T) {
	spec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Memory Leak Test", "version": "1.0.0"},
		"components": {
			"schemas": {
				"LeakTestSchema": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"data": {"type": "string"}
					}
				}
			}
		}
	}`

	config := &Config{
		PackageName: "leaktest",
		OutputDir:   "./test",
	}

	// Force garbage collection and get initial memory stats
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run generation many times
	for i := 0; i < 100; i++ {
		gen := NewGenerator(config)
		_, err := gen.ParseAndGenerate([]byte(spec))
		if err != nil {
			t.Fatalf("Generation %d failed: %v", i, err)
		}
	}

	// Force garbage collection and check memory
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Use HeapAlloc for more accurate measurement and handle potential underflow
	memoryGrowth := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
	if memoryGrowth < 0 {
		memoryGrowth = 0 // Memory was actually freed
	}

	t.Logf("Memory growth after 100 generations: %d bytes", memoryGrowth)

	// Memory growth should be minimal (less than 10MB for 100 iterations)
	maxGrowthMB := float64(10)
	actualGrowthMB := float64(memoryGrowth) / (1024 * 1024)
	if actualGrowthMB > maxGrowthMB {
		t.Errorf("Potential memory leak detected: %.2f MB growth (max: %.2f MB)",
			actualGrowthMB, maxGrowthMB)
	}
}
