package main

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/urfave/cli/v3"
)

var (
	cpeGzipFileDonwloadURL = "https://nvd.nist.gov/feeds/xml/cpe/dictionary/official-cpe-dictionary_v2.3.xml.gz"
)

// downloadFile 下载文件到指定路径
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractGZ 解压 .gz 文件
func extractGZ(gzFilePath, outputFilePath string) error {
	gzFile, err := os.Open(gzFilePath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gzReader)
	return err
}

// streamReadXML 流式读取 XML 文件
func streamReadXML(xmlFilePath string) (<-chan CPEItem) {
	cpeItemCh := make(chan CPEItem, 1000)
	go func() {
		file, err := os.Open(xmlFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()

		decoder := xml.NewDecoder(file)
		count := 0

		defer close(cpeItemCh)
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
				continue
			}
	
			switch se := token.(type) {
			case xml.StartElement:
				if se.Name.Local == "cpe-item" {
					var item CPEItem
					if err := decoder.DecodeElement(&item, &se); err != nil {
						log.Println(err)
						continue
					}
	
					cpeItemCh <- item
	
					count += 1
				}
			}
		}
		fmt.Printf("parse cpe total %d\n", count)
	}()
	return cpeItemCh
}

func main() {
    cmd := &cli.Command{
        // Name:  "findout",
        Usage: "download cpe and import to database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "output-file",
				Usage: "database file path",
				Required: true,
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name: "type",
				Usage: "database type (duckdb, sqlite3)",
				Value: "duckdb",
				Aliases: []string{"t"},
				Action: func(ctx context.Context, cmd *cli.Command, v string) error {
					if !slices.Contains([]string{"duckdb", "sqlite3"}, v) {
						return fmt.Errorf("allowed values are duckdb, sqlite3")
					}
					return nil
				},
			},
		},
        Action: func(ctx context.Context, cmd *cli.Command) error {
			tmpDir, err := os.MkdirTemp("", "cpe-dictionary")
			if err != nil {
				return fmt.Errorf("failed to create temp directory: %w", err)
			}
			defer os.RemoveAll(tmpDir) // 清理临时文件夹
		
			gzFilePath := filepath.Join(tmpDir, "official-cpe-dictionary_v2.3.xml.gz")
			if err := downloadFile(cpeGzipFileDonwloadURL, gzFilePath); err != nil {
				return fmt.Errorf("failed to download file: %w", err)
			}
		
			// 解压文件
			xmlFilePath := filepath.Join(tmpDir, "official-cpe-dictionary_v2.3.xml")
			if err := extractGZ(gzFilePath, xmlFilePath); err != nil {
				return fmt.Errorf("failed to extract GZ file: %w", err)
			}
			var d DB
			if cmd.String("type") == "duckdb" {
				d, err = NewDuckDB(cmd.String("output-file"))
				if err != nil {
					return fmt.Errorf("NewDuckDB error: %w", err)
				}
			} else {
				d, err = NewSqliteDB(cmd.String("output-file"))
				if err != nil {
					return fmt.Errorf("NewSqliteDB error: %w", err)
				}
			}
			// 流式读取 XML 文件
			ch := make(chan CPE23, 100)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := d.WriteRows(ch)
				if err != nil {
					os.RemoveAll(tmpDir)
					log.Fatalf("err: %v", err)
				}
			}()
			for item := range streamReadXML(xmlFilePath) {
			// for item := range streamReadXML("/tmp/cpe-dictionary410080075/official-cpe-dictionary_v2.3.xml") {
				cpe23Data, err := ParseCPE23(item)
				if err != nil {
					log.Fatalf("err: %v", err)
				}
				ch <- *cpe23Data
			}
			close(ch)
			wg.Wait()
			return nil
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}