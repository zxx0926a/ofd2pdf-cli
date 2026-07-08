package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zc310/ofd/pkg/converter"
)

func main() {
	// 命令行参数
	inputPath := flag.String("i", "", "输入 OFD 文件路径（或目录，配合 -batch）")
	outputPath := flag.String("o", "", "输出 PDF 文件路径（或目录，配合 -batch）")
	dpi := flag.Int("dpi", 300, "输出分辨率 DPI")
	batch := flag.Bool("batch", false, "批量模式：-i 指定目录，转换目录下所有 .ofd 文件")
	workers := flag.Int("w", 4, "批量模式并发数")
	verbose := flag.Bool("v", false, "显示详细进度")
	help := flag.Bool("h", false, "显示帮助")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `ofd2pdf - OFD 转 PDF 命令行工具

用法:
  单文件转换:
    ofd2pdf.exe -i input.ofd -o output.pdf [-dpi 300]
  
  批量转换（保留目录结构）:
    ofd2pdf.exe -batch -i "输入目录" -o "输出目录" [-w 4] [-v]

参数:
`)
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), `
示例:
  ofd2pdf.exe -i 发票.ofd -o 发票.pdf
  ofd2pdf.exe -batch -i "D:\ofd_files" -o "D:\pdf_output" -w 8 -v
`)
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *inputPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "错误: 必须指定 -i 和 -o 参数")
		flag.Usage()
		os.Exit(1)
	}

	if *batch {
		runBatch(*inputPath, *outputPath, *dpi, *workers, *verbose)
	} else {
		runSingle(*inputPath, *outputPath, *dpi, *verbose)
	}
}

func runSingle(inputPath, outputPath string, dpi int, verbose bool) {
	if verbose {
		log.Printf("转换: %s -> %s (DPI: %d)", inputPath, outputPath, dpi)
	}

	// 检查输入文件
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Fatalf("输入文件不存在: %s", inputPath)
	}

	// 确保输出目录存在
	outDir := filepath.Dir(outputPath)
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Fatalf("创建输出目录失败: %v", err)
		}
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	// 转换
	err = converter.PDF(inputPath, outFile)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	if verbose {
		log.Printf("完成: %s", outputPath)
	}
}

func runBatch(inputDir, outputDir string, dpi, workers int, verbose bool) {
	// 检查输入目录
	info, err := os.Stat(inputDir)
	if err != nil {
		log.Fatalf("输入目录不存在: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("批量模式下 -i 必须是目录: %s", inputDir)
	}

	// 收集所有 .ofd 文件
	var files []string
	err = filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".ofd") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("遍历目录失败: %v", err)
	}

	if len(files) == 0 {
		log.Println("未找到任何 .ofd 文件")
		return
	}

	log.Printf("找到 %d 个 OFD 文件，开始转换（并发: %d）...", len(files), workers)

	// 并发转换
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	failed := 0

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(ofdPath string) {
			defer wg.Done()
			defer func() { <-sem }()

			// 计算相对路径
			relPath, _ := filepath.Rel(inputDir, ofdPath)
			// 替换扩展名为 .pdf
			pdfRel := strings.TrimSuffix(relPath, filepath.Ext(relPath)) + ".pdf"
			pdfPath := filepath.Join(outputDir, pdfRel)

			// 确保输出目录存在
			pdfDir := filepath.Dir(pdfPath)
			if err := os.MkdirAll(pdfDir, 0755); err != nil {
				mu.Lock()
				failed++
				if verbose {
					log.Printf("❌ %s: 创建目录失败: %v", ofdPath, err)
				}
				mu.Unlock()
				return
			}

			// 转换
			outFile, err := os.Create(pdfPath)
			if err != nil {
				mu.Lock()
				failed++
				if verbose {
					log.Printf("❌ %s: 创建输出失败: %v", ofdPath, err)
				}
				mu.Unlock()
				return
			}

			err = converter.PDF(ofdPath, outFile)
			outFile.Close()

			mu.Lock()
			if err != nil {
				failed++
				if verbose {
					log.Printf("❌ %s: %v", ofdPath, err)
				}
			} else {
				success++
				if verbose {
					log.Printf("✅ %s -> %s", ofdPath, pdfPath)
				}
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	log.Printf("完成: 成功 %d, 失败 %d", success, failed)
	if failed > 0 {
		os.Exit(1)
	}
}