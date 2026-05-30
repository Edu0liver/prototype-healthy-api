// Command seed inserts a demo tenant (company + admin + agent + KB + channel)
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
		return nil
	}); err != nil {
		log.Fatalf("seed company: %v", err)
	}

	if err := db.Tenant(ctx, companyID, func(ctx context.Context) error {
		tx := database.MustTx(ctx)
		hash, _ := hashPassword(demoPassword)
		stmts := []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO company_branding (company_id) VALUES (?) ON CONFLICT DO NOTHING`, []any{companyID}},
			{`INSERT INTO users (id,company_id,email,password_hash,name,role,status)
			  VALUES (?,?,?,?,'Demo Admin','admin','active')
			  ON CONFLICT (company_id,email) DO NOTHING`, []any{mustUUID(), companyID, demoEmail, hash}},
			{`INSERT INTO agents (id,company_id,name,system_prompt,model,temperature,max_output_tokens,handover_enabled,handover_keywords,status)
			  VALUES (?,?,'Demo Bot','Você é um atendente prestável da Demo Co.','gpt-4o-mini',0.7,1024,true,'["humano","atendente"]'::jsonb,'active')`,
				[]any{mustUUID(), companyID}},
			{`INSERT INTO knowledge_bases (id,company_id,name,embedding_model,chunk_size,chunk_overlap)
			  VALUES (?,?,'Demo KB','text-embedding-3-small',800,100)`, []any{mustUUID(), companyID}},
			{`INSERT INTO channels (id,company_id,type,name,status,metadata)
			  VALUES (?,?,'instagram','Demo IG','disconnected','{}'::jsonb)`, []any{mustUUID(), companyID}},
		}
		for _, s := range stmts {
			if err := tx.Exec(s.sql, s.args...).Error; err != nil {
				return err
			}
		}
		return nil
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
