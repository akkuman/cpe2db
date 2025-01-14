package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/marcboeker/go-duckdb"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB interface {
	WriteRows(<-chan CPE23) error
}

type SqliteDB struct {
	db *gorm.DB
}

func NewSqliteDB(path string) (DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&CPE23{})
	db.Exec("PRAGMA journal_mode = OFF;")
	db.Exec("PRAGMA synchronous = OFF;")
	// 设置内存映射大小为 256MB
	db.Exec("PRAGMA mmap_size=268435456;")
	db.Exec("PRAGMA cache_size=2000;")
	return &SqliteDB{
		db: db,
	}, nil
}

func (d *SqliteDB) WriteRows(ch <-chan CPE23) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	items := make([]CPE23, 0)
	count := 0
	Loop:
	for {
		select {
		case item, ok := <-ch:
			if !ok {
				break Loop
			}
			count += 1
			items = append(items, item)
			if len(items) == 10000 {
				tx := d.db.CreateInBatches(items, 2000)
				if tx.Error != nil {
					return tx.Error
				}
				items = make([]CPE23, 0) 
				log.Printf("已插入 %d 条", count)
			}
		case <- ticker.C:
			tx := d.db.CreateInBatches(items, 1000)
			if tx.Error != nil {
				return tx.Error
			}
			items = make([]CPE23, 0) 
		}
	}
	if len(items) > 0 {
		tx := d.db.CreateInBatches(items, 1000)
		if tx.Error == nil {
			log.Printf("已插入 %d 条", count)
		}
		return tx.Error
	}
	return nil
}

type DuckDB struct {
	connector *duckdb.Connector
}

func NewDuckDB(path string) (DB, error) {
	connector, err := duckdb.NewConnector(path, nil)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS cpe23 (
		"cpe_ver" VARCHAR,
		"title" VARCHAR,
		"part" VARCHAR,
		"vendor" VARCHAR,
		"product" VARCHAR,
		"version" VARCHAR,
		"update" VARCHAR,
		"edition" VARCHAR,
		"language" VARCHAR,
		"sw_edition" VARCHAR,
		"target_sw" VARCHAR,
		"target_hw" VARCHAR,
		"other" VARCHAR,
		"references" VARCHAR
	)`); err != nil {
		return nil, err
	}
	return &DuckDB{
		connector: connector,
	}, nil
}

func (d *DuckDB) WriteRows(ch <-chan CPE23) error {
	conn, err := d.connector.Connect(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	appender, err := duckdb.NewAppenderFromConn(conn, "", "cpe23")
	if err != nil {
		return err
	}
	defer appender.Close()
	count := 0
	for r := range ch {
		referencesJSON, _ := json.Marshal(r.References)
		err = appender.AppendRow(
			r.CPEVer,
			r.Title,
			r.Part,
			r.Vendor,
			r.Product,
			r.Version,
			r.Update,
			r.Edition,
			r.Language,
			r.SwEdition,
			r.TargetSw,
			r.TargetHw,
			r.Other,
			referencesJSON,
		)
		if err != nil {
			log.Printf("[ERR] %v", err)
			continue
		}
		count += 1
		if count % 10000 == 0 {
			log.Printf("已插入 %d 条", count)
		}
	}
	log.Printf("插入完成，已插入 %d 条", count)
	return nil
}