package storage

import (
	"context"
	"testing"
	"time"

	"L3_6/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB creates a PostgreSQL test container and applies migrations
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create pgxpool connection
	cfg, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)

	dbPool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	// Run migrations
	exitCode, _, err := postgresContainer.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-c", `
		CREATE TABLE IF NOT EXISTS sales (
			id SERIAL PRIMARY KEY,
			type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
			amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
			date TIMESTAMPTZ NOT NULL,
			category VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(date);
		CREATE INDEX IF NOT EXISTS idx_sales_category ON sales(category);
	`})
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	// Cleanup function
	cleanup := func() {
		dbPool.Close()
		postgresContainer.Terminate(ctx)
	}

	return dbPool, cleanup
}

var testSales = []models.Sale{
	{
		Type:     "income",
		Amount:   1000.50,
		Date:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Category: "Salary",
	},
	{
		Type:     "expense",
		Amount:   250.75,
		Date:     time.Date(2024, 1, 16, 14, 15, 0, 0, time.UTC),
		Category: "Food",
	},
	{
		Type:     "expense",
		Amount:   1200.00,
		Date:     time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC),
		Category: "Rent",
	},
	{
		Type:     "income",
		Amount:   500.00,
		Date:     time.Date(2024, 1, 18, 16, 45, 0, 0, time.UTC),
		Category: "Freelance",
	},
}

func TestStorage_CreateSale(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	t.Run("create valid sale", func(t *testing.T) {
		sale := testSales[0]
		err := storage.CreateSale(&sale)
		require.NoError(t, err)
		assert.Equal(t, 1, sale.ID)

		// Verify the sale was created
		var retrievedSale models.Sale
		err = db.QueryRow(context.Background(),
			"SELECT id, type, amount, date, category FROM sales WHERE id = $1",
			sale.ID).
			Scan(&retrievedSale.ID, &retrievedSale.Type, &retrievedSale.Amount,
				&retrievedSale.Date, &retrievedSale.Category)
		require.NoError(t, err)
		assert.Equal(t, sale.Type, retrievedSale.Type)
		assert.Equal(t, sale.Amount, retrievedSale.Amount)
		assert.Equal(t, sale.Category, retrievedSale.Category)
		assert.WithinDuration(t, sale.Date, retrievedSale.Date, time.Second)
	})

	t.Run("create multiple sales", func(t *testing.T) {
		for i, testSale := range testSales[1:] {
			sale := testSale
			err := storage.CreateSale(&sale)
			require.NoError(t, err)
			assert.Equal(t, i+2, sale.ID) // ID should be sequential
		}
	})

	t.Run("create sale with zero amount should fail", func(t *testing.T) {
		invalidSale := models.Sale{
			Type:     "expense",
			Amount:   0.00,
			Date:     time.Now(),
			Category: "Test",
		}
		err := storage.CreateSale(&invalidSale)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage.CreateSale")
	})
}

func TestStorage_GetSales(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	t.Run("get empty sales", func(t *testing.T) {
		sales, err := storage.GetSales()
		require.NoError(t, err)
		assert.Empty(t, sales)
	})

	t.Run("get sales ordered by date desc", func(t *testing.T) {
		// Create test data
		for _, testSale := range testSales {
			sale := testSale
			err := storage.CreateSale(&sale)
			require.NoError(t, err)
		}

		sales, err := storage.GetSales()
		require.NoError(t, err)
		require.Len(t, sales, len(testSales))

		// Verify ordering (most recent first)
		for i := 1; i < len(sales); i++ {
			assert.True(t, sales[i-1].Date.After(sales[i].Date) || sales[i-1].Date.Equal(sales[i].Date),
				"Sales should be ordered by date DESC")
		}
	})

	t.Run("verify all sales data", func(t *testing.T) {
		// Clear existing data and create fresh
		db.Exec(context.Background(), "DELETE FROM sales")

		createdSales := make([]models.Sale, len(testSales))
		for i, testSale := range testSales {
			sale := testSale
			err := storage.CreateSale(&sale)
			require.NoError(t, err)
			createdSales[i] = sale
		}

		sales, err := storage.GetSales()
		require.NoError(t, err)
		assert.Len(t, sales, len(testSales))

		// Verify each sale matches
		for _, retrievedSale := range sales {
			var expected models.Sale
			for _, original := range createdSales {
				if original.ID == retrievedSale.ID {
					expected = original
					break
				}
			}
			assert.Equal(t, expected.Type, retrievedSale.Type)
			assert.Equal(t, expected.Amount, retrievedSale.Amount)
			assert.Equal(t, expected.Category, retrievedSale.Category)
			assert.WithinDuration(t, expected.Date, retrievedSale.Date, time.Second)
		}
	})
}

func TestStorage_UpdateSale(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	// Create a sale first
	sale := testSales[0]
	err := storage.CreateSale(&sale)
	require.NoError(t, err)
	originalID := sale.ID

	t.Run("update existing sale", func(t *testing.T) {
		sale.Type = "expense"
		sale.Amount = 750.25
		sale.Category = "Updated Category"
		sale.Date = time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)

		err := storage.UpdateSale(&sale)
		require.NoError(t, err)
		assert.Equal(t, originalID, sale.ID) // ID should remain unchanged

		// Verify the update
		var retrievedSale models.Sale
		err = db.QueryRow(context.Background(),
			"SELECT id, type, amount, date, category FROM sales WHERE id = $1",
			sale.ID).
			Scan(&retrievedSale.ID, &retrievedSale.Type, &retrievedSale.Amount,
				&retrievedSale.Date, &retrievedSale.Category)
		require.NoError(t, err)
		assert.Equal(t, sale.Type, retrievedSale.Type)
		assert.Equal(t, sale.Amount, retrievedSale.Amount)
		assert.Equal(t, sale.Category, retrievedSale.Category)
		assert.WithinDuration(t, sale.Date, retrievedSale.Date, time.Second)
	})

	t.Run("update non-existent sale", func(t *testing.T) {
		nonExistentSale := models.Sale{
			ID:       999,
			Type:     "income",
			Amount:   100.0,
			Date:     time.Now(),
			Category: "Test",
		}
		err := storage.UpdateSale(&nonExistentSale)
		require.NoError(t, err) // UPDATE without WHERE match doesn't error in PostgreSQL
	})
}

func TestStorage_DeleteSale(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	t.Run("delete existing sale", func(t *testing.T) {
		// Create a sale
		sale := testSales[0]
		err := storage.CreateSale(&sale)
		require.NoError(t, err)

		// Verify it exists
		var count int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM sales WHERE id = $1", sale.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Delete it
		err = storage.DeleteSale(sale.ID)
		require.NoError(t, err)

		// Verify it's gone
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM sales WHERE id = $1", sale.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete non-existent sale", func(t *testing.T) {
		err := storage.DeleteSale(999)
		require.NoError(t, err) // DELETE without match doesn't error in PostgreSQL
	})
}

func TestStorage_GetAnalytics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	t.Run("analytics with no data", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC)

		analytics, err := storage.GetAnalytics(from, to)
		require.NoError(t, err)
		assert.Equal(t, 0.0, analytics.Sum)
		assert.Equal(t, 0.0, analytics.Average)
		assert.Equal(t, 0, analytics.Count)
		assert.Equal(t, 0.0, analytics.Median)
		assert.Equal(t, 0.0, analytics.Percentile90)
	})

	t.Run("analytics with test data", func(t *testing.T) {
		// Create test sales
		for _, testSale := range testSales {
			sale := testSale
			err := storage.CreateSale(&sale)
			require.NoError(t, err)
		}

		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC)

		analytics, err := storage.GetAnalytics(from, to)
		require.NoError(t, err)

		// Expected: sum = 2450.5, count = 4, average = 612.625
		expectedSum := 1000.50 + 250.75 + 1200.00 + 500.00
		assert.Equal(t, expectedSum, analytics.Sum)
		assert.Equal(t, 4, analytics.Count)
		assert.Equal(t, expectedSum/4.0, analytics.Average)
		assert.NotZero(t, analytics.Median)
		assert.NotZero(t, analytics.Percentile90)
	})

	t.Run("analytics with date range filter", func(t *testing.T) {
		// Clear and recreate data for this test
		db.Exec(context.Background(), "DELETE FROM sales")

		// Create sales in different months
		janSale := models.Sale{
			Type:     "income",
			Amount:   1000.0,
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Category: "January",
		}
		err := storage.CreateSale(&janSale)
		require.NoError(t, err)

		febSale := models.Sale{
			Type:     "income",
			Amount:   2000.0,
			Date:     time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Category: "February",
		}
		err = storage.CreateSale(&febSale)
		require.NoError(t, err)

		// Filter for January only
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)

		analytics, err := storage.GetAnalytics(from, to)
		require.NoError(t, err)
		assert.Equal(t, 1000.0, analytics.Sum)
		assert.Equal(t, 1, analytics.Count)

		// Filter for February only
		from = time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		to = time.Date(2024, 2, 28, 23, 59, 59, 999999999, time.UTC)

		analytics, err = storage.GetAnalytics(from, to)
		require.NoError(t, err)
		assert.Equal(t, 2000.0, analytics.Sum)
		assert.Equal(t, 1, analytics.Count)
	})

	t.Run("analytics statistical calculations", func(t *testing.T) {
		// Clear and create sales for statistical testing
		db.Exec(context.Background(), "DELETE FROM sales")

		// Create sales with predictable values for median/percentile testing
		testValues := []float64{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0}
		for i, amount := range testValues {
			sale := models.Sale{
				Type:     "income",
				Amount:   amount,
				Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
				Category: "Statistical Test",
			}
			err := storage.CreateSale(&sale)
			require.NoError(t, err)
		}

		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)

		analytics, err := storage.GetAnalytics(from, to)
		require.NoError(t, err)

		assert.Equal(t, 550.0, analytics.Sum)
		assert.Equal(t, 10, analytics.Count)
		assert.Equal(t, 55.0, analytics.Average)
		assert.Equal(t, 55.0, analytics.Median)       // Median should be 55 for 10 values
		assert.Equal(t, 91.0, analytics.Percentile90) // 90th percentile for this data
	})
}

func TestStorage_ErrorHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := NewStorage(db)

	t.Run("invalid type constraint", func(t *testing.T) {
		invalidSale := models.Sale{
			Type:     "invalid", // Long enough to exceed varchar(10) limit
			Amount:   100.0,
			Date:     time.Now(),
			Category: "Test",
		}

		err := storage.CreateSale(&invalidSale)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage.CreateSale")
	})

	t.Run("invalid type CHECK constraint", func(t *testing.T) {
		// Test using raw SQL to bypass type validation
		_, err := db.Exec(context.Background(),
			"INSERT INTO sales (type, amount, date, category) VALUES ($1, $2, $3, $4)",
			"invalid", 100.0, time.Now(), "Test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "check constraint")
	})

	t.Run("negative amount constraint", func(t *testing.T) {
		// Test using raw SQL to bypass validation
		_, err := db.Exec(context.Background(),
			"INSERT INTO sales (type, amount, date, category) VALUES ($1, $2, $3, $4)",
			"income", -100.0, time.Now(), "Test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "check constraint")
	})
}
