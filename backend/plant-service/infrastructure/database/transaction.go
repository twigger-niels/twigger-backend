package database

import (
	"context"
	"database/sql"
	"fmt"
)

// TxManager manages database transactions
type TxManager struct {
	db *sql.DB
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is always finalized
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred - rollback and re-panic
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		// Function returned error - rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Success - commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransactionIsolation executes a function within a transaction with specific isolation level
func (tm *TxManager) WithTransactionIsolation(
	ctx context.Context,
	isolationLevel sql.IsolationLevel,
	fn func(*sql.Tx) error,
) error {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: isolationLevel,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TxRepository wraps repositories to work within transactions
type TxRepository struct {
	tx *sql.Tx
}

// NewTxRepository creates a repository wrapper for transaction operations
func NewTxRepository(tx *sql.Tx) *TxRepository {
	return &TxRepository{tx: tx}
}

// QueryRowContext executes a query within the transaction
func (r *TxRepository) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.tx.QueryRowContext(ctx, query, args...)
}

// QueryContext executes a query within the transaction
func (r *TxRepository) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.tx.QueryContext(ctx, query, args...)
}

// ExecContext executes a command within the transaction
func (r *TxRepository) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.tx.ExecContext(ctx, query, args...)
}

// --- Example Usage Patterns ---

// Example 1: Simple transaction with multiple operations
// func (s *PlantService) CreatePlantWithGrowingConditions(
//     ctx context.Context,
//     plant *entity.Plant,
//     conditions *types.GrowingConditions,
// ) error {
//     return s.txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
//         // Create plant
//         _, err := tx.ExecContext(ctx, "INSERT INTO plants ...", plant.PlantID, ...)
//         if err != nil {
//             return fmt.Errorf("failed to create plant: %w", err)
//         }
//
//         // Create growing conditions
//         _, err = tx.ExecContext(ctx, "INSERT INTO growing_conditions_assertions ...", ...)
//         if err != nil {
//             return fmt.Errorf("failed to create growing conditions: %w", err)
//         }
//
//         return nil // Commit
//     })
// }

// Example 2: Transaction with read-write isolation
// func (s *PlantService) UpdatePlantInventory(ctx context.Context, plantID string, delta int) error {
//     return s.txManager.WithTransactionIsolation(
//         ctx,
//         sql.LevelSerializable,
//         func(tx *sql.Tx) error {
//             // Read current inventory
//             var count int
//             err := tx.QueryRowContext(ctx, "SELECT inventory_count FROM plants WHERE plant_id = $1", plantID).Scan(&count)
//             if err != nil {
//                 return err
//             }
//
//             // Update inventory
//             newCount := count + delta
//             if newCount < 0 {
//                 return fmt.Errorf("insufficient inventory")
//             }
//
//             _, err = tx.ExecContext(ctx, "UPDATE plants SET inventory_count = $1 WHERE plant_id = $2", newCount, plantID)
//             return err
//         },
//     )
// }

// Example 3: Multi-repository transaction
// type TransactionalService struct {
//     db        *sql.DB
//     txManager *TxManager
// }
//
// func (s *TransactionalService) TransferPlantOwnership(
//     ctx context.Context,
//     plantID, fromGardenID, toGardenID string,
// ) error {
//     return s.txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
//         txRepo := NewTxRepository(tx)
//
//         // Remove from old garden
//         _, err := txRepo.ExecContext(ctx,
//             "DELETE FROM garden_plants WHERE garden_id = $1 AND plant_id = $2",
//             fromGardenID, plantID)
//         if err != nil {
//             return fmt.Errorf("failed to remove from old garden: %w", err)
//         }
//
//         // Add to new garden
//         _, err = txRepo.ExecContext(ctx,
//             "INSERT INTO garden_plants (garden_id, plant_id, planted_at) VALUES ($1, $2, NOW())",
//             toGardenID, plantID)
//         if err != nil {
//             return fmt.Errorf("failed to add to new garden: %w", err)
//         }
//
//         // Update plant history
//         _, err = txRepo.ExecContext(ctx,
//             "INSERT INTO plant_history (plant_id, event, from_garden, to_garden) VALUES ($1, 'transfer', $2, $3)",
//             plantID, fromGardenID, toGardenID)
//         if err != nil {
//             return fmt.Errorf("failed to record history: %w", err)
//         }
//
//         return nil
//     })
// }

// SavepointManager manages savepoints within a transaction
type SavepointManager struct {
	tx *sql.Tx
}

// NewSavepointManager creates a savepoint manager for the transaction
func NewSavepointManager(tx *sql.Tx) *SavepointManager {
	return &SavepointManager{tx: tx}
}

// CreateSavepoint creates a named savepoint
func (sm *SavepointManager) CreateSavepoint(ctx context.Context, name string) error {
	_, err := sm.tx.ExecContext(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}
	return nil
}

// RollbackToSavepoint rolls back to a named savepoint
func (sm *SavepointManager) RollbackToSavepoint(ctx context.Context, name string) error {
	_, err := sm.tx.ExecContext(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", name, err)
	}
	return nil
}

// ReleaseSavepoint releases a named savepoint
func (sm *SavepointManager) ReleaseSavepoint(ctx context.Context, name string) error {
	_, err := sm.tx.ExecContext(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", name, err)
	}
	return nil
}

// WithSavepoint executes a function within a savepoint
// If the function fails, rolls back to the savepoint
func (sm *SavepointManager) WithSavepoint(ctx context.Context, name string, fn func() error) error {
	if err := sm.CreateSavepoint(ctx, name); err != nil {
		return err
	}

	if err := fn(); err != nil {
		if rbErr := sm.RollbackToSavepoint(ctx, name); rbErr != nil {
			return fmt.Errorf("operation failed: %w, savepoint rollback failed: %v", err, rbErr)
		}
		return err
	}

	return sm.ReleaseSavepoint(ctx, name)
}
