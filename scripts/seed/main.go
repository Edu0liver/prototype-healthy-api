// Command seed inserts a minimal demo tenant (company + admin user)
// for local development. Run with: make seed (DB must be migrated).
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	demoSlug     = "demo"
	demoEmail    = "admin@demo.com"
	demoPassword = "demodemo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	gdb, err := gorm.Open(gormpg.Open(cfg.Database.URL), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	db := &database.DB{DB: gdb}
	ctx := context.Background()

	companyID := mustUUID()
	if err := db.System(ctx, func(ctx context.Context) error {
		tx := database.MustTx(ctx)
		if err := tx.Exec(
			`INSERT INTO companies (id,name,slug,status,plan) VALUES (?,?,?,'active','pro')
			 ON CONFLICT (slug) DO NOTHING`, companyID, "Demo Co", demoSlug).Error; err != nil {
			return err
		}
		// If the company already existed, fetch its id (scan as text then parse).
		var idStr string
		if err := tx.Raw(`SELECT id::text FROM companies WHERE slug = ?`, demoSlug).Scan(&idStr).Error; err != nil {
			return err
		}
		companyID = uuid.MustParse(idStr)

		// Map dev hosts -> demo tenant so Host->tenant resolution (branding,
		// /branding/host) works locally. The frontend Host includes the port.
		for _, domain := range []string{"localhost:3000", "localhost"} {
			if err := tx.Exec(
				`INSERT INTO company_domains (id,company_id,domain,is_primary,verified_at)
				 VALUES (?,?,?,?,now()) ON CONFLICT (domain) DO NOTHING`,
				mustUUID(), companyID, domain, domain == "localhost:3000").Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("seed company: %v", err)
	}

	if err := db.Tenant(ctx, companyID, func(ctx context.Context) error {
		tx := database.MustTx(ctx)
		hash, _ := hashPassword(demoPassword)
		if err := tx.Exec(`INSERT INTO company_branding (id,company_id) VALUES (?,?) ON CONFLICT (company_id) DO NOTHING`,
			mustUUID(), companyID).Error; err != nil {
			return err
		}
		return tx.Exec(`INSERT INTO users (id,company_id,email,password_hash,name,role_id,status)
			  VALUES (?,?,?,?,'Admin',(SELECT id FROM roles WHERE name='admin'),'active')
			  ON CONFLICT (email) DO NOTHING`,
			mustUUID(), companyID, demoEmail, hash).Error
	}); err != nil {
		log.Fatalf("seed tenant data: %v", err)
	}

	fmt.Printf("Seeded demo tenant:\n  company slug: %s\n  admin:        %s / %s\n", demoSlug, demoEmail, demoPassword)
}

func mustUUID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

// hashPassword mirrors iam's argon2id encoding so the seeded admin can log in.
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	const mem, t, threads, keyLen = 64 * 1024, 1, 4, 32
	key := argon2.IDKey([]byte(password), salt, t, mem, threads, keyLen)
	return strings.Join([]string{
		"", "argon2id",
		fmt.Sprintf("v=%d", argon2.Version),
		fmt.Sprintf("m=%d,t=%d,p=%d", mem, t, threads),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	}, "$"), nil
}
