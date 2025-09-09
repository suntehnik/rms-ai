package benchmarks

import (
	"context"
	"testing"

	"product-requirements-management/internal/benchmarks/setup"
)

// BenchmarkDataGeneration tests the performance of data generation utilities
func BenchmarkDataGeneration(b *testing.B) {
	ctx := context.Background()
	
	// Create PostgreSQL container
	dbContainer, err := setup.NewPostgreSQLContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create PostgreSQL container: %v", err)
	}
	defer dbContainer.Cleanup(ctx)

	// Create data generator
	dataGen := setup.NewDataGenerator(dbContainer.DB)

	b.Run("SmallDataSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Reset database before each iteration
			if err := dbContainer.ResetDatabase(); err != nil {
				b.Fatalf("Failed to reset database: %v", err)
			}

			// Generate small dataset
			config := setup.GetSmallDataSet()
			if err := dataGen.GenerateFullDataSet(config); err != nil {
				b.Fatalf("Failed to generate small dataset: %v", err)
			}
		}
	})

	b.Run("MediumDataSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Reset database before each iteration
			if err := dbContainer.ResetDatabase(); err != nil {
				b.Fatalf("Failed to reset database: %v", err)
			}

			// Generate medium dataset
			config := setup.GetMediumDataSet()
			if err := dataGen.GenerateFullDataSet(config); err != nil {
				b.Fatalf("Failed to generate medium dataset: %v", err)
			}
		}
	})

	b.Run("UserCreation", func(b *testing.B) {
		// Reset database once
		if err := dbContainer.ResetDatabase(); err != nil {
			b.Fatalf("Failed to reset database: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create 100 users
			_, err := dataGen.CreateUsers(100)
			if err != nil {
				b.Fatalf("Failed to create users: %v", err)
			}
			
			// Cleanup after each iteration
			if err := dataGen.CleanupDatabase(); err != nil {
				b.Fatalf("Failed to cleanup database: %v", err)
			}
		}
	})

	b.Run("EpicCreation", func(b *testing.B) {
		// Reset database and create users once
		if err := dbContainer.ResetDatabase(); err != nil {
			b.Fatalf("Failed to reset database: %v", err)
		}
		
		users, err := dataGen.CreateUsers(50)
		if err != nil {
			b.Fatalf("Failed to create users: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create 100 epics
			_, err := dataGen.CreateEpics(100, users)
			if err != nil {
				b.Fatalf("Failed to create epics: %v", err)
			}
			
			// Cleanup epics after each iteration
			if err := dbContainer.DB.Exec("DELETE FROM epics").Error; err != nil {
				b.Fatalf("Failed to cleanup epics: %v", err)
			}
		}
	})

	b.Run("BulkDataInsertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Reset database before each iteration
			if err := dbContainer.ResetDatabase(); err != nil {
				b.Fatalf("Failed to reset database: %v", err)
			}

			// Test bulk insertion performance
			users, err := dataGen.CreateUsers(1000)
			if err != nil {
				b.Fatalf("Failed to create users: %v", err)
			}

			epics, err := dataGen.CreateEpics(500, users)
			if err != nil {
				b.Fatalf("Failed to create epics: %v", err)
			}

			userStories, err := dataGen.CreateUserStories(3, epics, users)
			if err != nil {
				b.Fatalf("Failed to create user stories: %v", err)
			}

			_, err = dataGen.CreateRequirements(2, userStories, users)
			if err != nil {
				b.Fatalf("Failed to create requirements: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseOperations tests database cleanup and reset performance
func BenchmarkDatabaseOperations(b *testing.B) {
	ctx := context.Background()
	
	// Create PostgreSQL container
	dbContainer, err := setup.NewPostgreSQLContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create PostgreSQL container: %v", err)
	}
	defer dbContainer.Cleanup(ctx)

	dataGen := setup.NewDataGenerator(dbContainer.DB)

	b.Run("DatabaseCleanup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Generate some data first
			config := setup.GetSmallDataSet()
			if err := dataGen.GenerateFullDataSet(config); err != nil {
				b.Fatalf("Failed to generate dataset: %v", err)
			}

			// Measure cleanup performance
			b.StartTimer()
			if err := dataGen.CleanupDatabase(); err != nil {
				b.Fatalf("Failed to cleanup database: %v", err)
			}
			b.StopTimer()
		}
	})

	b.Run("DatabaseReset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Generate some data first
			config := setup.GetSmallDataSet()
			if err := dataGen.GenerateFullDataSet(config); err != nil {
				b.Fatalf("Failed to generate dataset: %v", err)
			}

			// Measure reset performance
			b.StartTimer()
			if err := dataGen.ResetDatabase(); err != nil {
				b.Fatalf("Failed to reset database: %v", err)
			}
			b.StopTimer()
		}
	})
}