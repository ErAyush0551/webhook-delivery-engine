package database

import (
	"fmt"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"    
)

func  RunMigrations(ctx context.Context, db *pgxpool.Pool)error {
	migrationDir := "migration"

	files ,err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	sort.Slice(files , func(i, j int) bool {
		return files[i].Name()<files[j].Name()	
	})

	for _, file := range files{
		if file.IsDir() || !strings.HasSuffix(file.Name (),".sql"){
			continue
		}
			path := filepath.Join(migrationDir, file.Name())

			sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		if _, err := db.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file.Name(), err)
		}

		fmt.Println("Migration applied:", file.Name())



	}


	return nil


}