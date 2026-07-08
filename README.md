# ofd2pdf-cli

基于 [zc310/ofd](https://github.com/zc310/ofd) 构建的 OFD 转 PDF 命令行工具。

## 特性

- ✅ 单文件 EXE，无依赖，开箱即用
- ✅ 支持中文路径、空格、特殊字符
- ✅ 支持单文件转换和批量转换（保持目录结构）
- ✅ 可自定义输出分辨率 (DPI)
- ✅ 跨平台：Windows / Linux / macOS (x64 / ARM64)
- ✅ 转换失败自动清理残留文件

## 下载

去 [Releases](https://github.com/zxx/ofd2pdf-cli/releases) 下载最新版 `ofd2pdf-windows-amd64.exe`（约 5-8 MB）。

## 用法

```cmd
REM 单文件转换（输出同名 .pdf）
ofd2pdf.exe "输入文件.ofd"

REM 指定输出路径
ofd2pdf.exe "输入文件.ofd" "输出文件.pdf"

REM 使用参数
ofd2pdf.exe -i "输入文件.ofd" -o "输出文件.pdf" -dpi 300

REM 

REM 批量转换整个目录（保持子目录结构）
ofd2pdf.exe -batch -i "C:\ofd文件夹" -o "C:\pdf输出"
```

## 参数说明

| 参数 | 短参 | 说明 |
|------|------|------|
| `-i` | | 输入 OFD 文件路径（或目录，配合 `-batch`） |
| `-o` | | 输出 PDF 文件路径（或目录，配合 `-batch`） |
| `-dpi` | | 输出分辨率，默认 300 |
| `-batch` | | 批量模式：`-i` 指定目录，转换目录下所有 `.ofd` 文件 |

## 批量转换示例

目录结构：
```
输入目录/
├── 文件1.ofd
├── 子目录/
│   └── 文件2.ofd
```

运行：
```cmd
ofd2pdf.exe -batch -i "输入目录" -o "输出目录"
```

输出结构：
```
输出目录/
├── 文件1.pdf
├── 子目录/
│   └── 文件2.pdf
```

## 构建

```bash
git clone https://github.com/zxx/ofd2pdf-cli.git
cd ofd2pdf-cli
go build -ldflags="-s -w" -o ofd2pdf.exe .
```

## 许可证

MIT License - 基于 zc310/ofd (MIT)