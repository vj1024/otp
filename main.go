package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v3"
)

// Options 定义命令行选项
type Options struct {
	Secret   string `short:"s" long:"secret" description:"TOTP Secret (base32 编码)" required:"true"`
	Config   string `short:"c" long:"config" description:"Secret 配置文件" default:"~/.config/otp/otp.yaml"`
	Verbose  bool   `short:"v" long:"verbose" description:"详细模式，显示所有信息"`
	Interval int    `short:"i" long:"interval" description:"自动刷新间隔（秒），0表示不自动刷新" default:"0"`
	Version  bool   `long:"version" description:"显示版本信息"`
	QRCode   bool   `short:"q" long:"qrcode" description:"显示二维码URI（用于添加验证器）"`
	Issuer   string `long:"issuer" description:"发行者名称（用于生成二维码URI）" default:"TOTP Generator"`
	Account  string `short:"a" long:"account" description:"账户名称（用于生成二维码URI）" default:"user@example.com"`
	Help     bool   `short:"h" long:"help" description:"显示帮助信息"`
}

// 版本信息
const (
	Version = "1.0.0"
	AppName = "TOTP Generator"
)

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "otp"
	parser.Usage = "[-s SECRET] [选项]"
	parser.LongDescription = "TOTP 代码生成器，兼容 Google Authenticator\n" +
		"支持生成6位验证码，每30秒自动刷新"

	// 解析命令行参数
	_, err := parser.Parse()
	if err != nil {
		// 如果错误是帮助请求，正常退出
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		// 如果错误是必需的参数缺失，显示帮助
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrRequired {
			fmt.Printf("错误: %v\n\n", err)
			parser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 显示版本信息
	if opts.Version {
		showVersion()
		os.Exit(0)
	}

	// 显示帮助信息
	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// 从配置读取 Secret
	secret := loadSecret(opts.Secret, opts.Config)

	// 清理 Secret
	secret = cleanSecret(secret)
	if secret == "" {
		fmt.Fprintf(os.Stderr, "错误: Secret 不能为空\n")
		os.Exit(1)
	}

	// 验证 Secret 格式
	if !isValidSecret(secret) {
		fmt.Fprintf(os.Stderr, "错误: Secret 不是有效的 base32 编码: `%s`\n", secret)
		os.Exit(1)
	}

	// 生成当前 TOTP
	code, err := generateTOTP(secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法生成 TOTP - %v\n", err)
		os.Exit(1)
	}

	// 如果需要自动刷新
	if opts.Interval > 0 {
		autoRefreshTOTP(secret, opts.Interval, opts.Verbose)
		return
	}

	// 根据选项输出
	if opts.Verbose {
		displayVerbose(code, secret, &opts)
	} else {
		// 简洁模式：只输出 code
		fmt.Println(code)
	}
}

func loadSecret(secret, filePath string) string {
	if strings.HasPrefix(filePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return secret
		}
		filePath = filepath.Join(home, filePath[2:])
	}

	bs, err := os.ReadFile(filePath)
	if err != nil {
		return secret
	}

	mp := make(map[string]string)
	if err := yaml.Unmarshal(bs, &mp); err != nil {
		return secret
	}

	if v, ok := mp[secret]; ok && v != "" {
		return v
	}
	return secret
}

func cleanSecret(secret string) string {
	// 移除所有空格
	secret = strings.ReplaceAll(secret, " ", "")
	// 转换为大写
	secret = strings.ToUpper(secret)
	return secret
}

func isValidSecret(secret string) bool {
	// 简单的 base32 验证：只包含 A-Z 和 2-7
	for _, ch := range secret {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '2' && ch <= '7')) {
			return false
		}
	}
	return len(secret) >= 16 // 通常至少 16 个字符
}

func generateTOTP(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func formatSecret(secret string) string {
	var result strings.Builder
	for i, ch := range secret {
		result.WriteRune(ch)
		if (i+1)%4 == 0 && i != len(secret)-1 {
			result.WriteString(" ")
		}
	}
	return result.String()
}

func displayVerbose(code, secret string, opts *Options) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📱 %s\n", AppName)
	fmt.Println(strings.Repeat("-", 50))

	currentTime := time.Now()
	remainingSeconds := 30 - int(currentTime.Unix()%30)

	fmt.Printf("🔢 当前代码: \033[1;32m%s\033[0m\n", code)
	fmt.Printf("⏱️  当前时间: %s\n", currentTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("⏳ 剩余时间: %d 秒\n", remainingSeconds)
	displayProgressBar(remainingSeconds)

	if opts.QRCode {
		fmt.Printf("📱 二维码URI: otpauth://totp/%s:%s?secret=%s&issuer=%s\n",
			opts.Issuer, opts.Account, secret, opts.Issuer)
	}

	fmt.Println(strings.Repeat("=", 50))
}

func displayProgressBar(remainingSeconds int) {
	totalWidth := 20
	filledWidth := int(float64(totalWidth) * (float64(30-remainingSeconds) / 30))

	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < totalWidth; i++ {
		if i < filledWidth {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	bar.WriteString("]")

	fmt.Printf("📊 进度: %s\n", bar.String())
}

func autoRefreshTOTP(secret string, interval int, verbose bool) {
	lastCode := ""
	for i := 0; ; i++ {
		if i > 0 {
			time.Sleep(time.Duration(interval) * time.Second)
		}

		code, err := generateTOTP(secret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "刷新错误: %v\n", err)
			continue
		}

		if verbose {
			// 详细模式：清屏并重新显示所有信息
			fmt.Print("\033[2J\033[H") // 清屏
			currentTime := time.Now()
			remainingSeconds := 30 - int(currentTime.Unix()%30)

			fmt.Println(strings.Repeat("=", 50))
			fmt.Printf("📱 %s (自动刷新)\n", AppName)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("🔢 当前代码: \033[1;32m%s\033[0m\n", code)
			fmt.Printf("⏱️  当前时间: %s\n", currentTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("⏳ 剩余时间: %d 秒\n", remainingSeconds)
			displayProgressBar(remainingSeconds)
			fmt.Printf("⏰ 刷新间隔: %d 秒\n", interval)
			fmt.Println(strings.Repeat("=", 50))
			fmt.Println("\n🔄 自动刷新中... 按 Ctrl+C 退出")
		} else if code != lastCode {
			// 简洁模式：只有当代码变化时才输出
			fmt.Println(code)
			lastCode = code
		}
	}
}

func showVersion() {
	fmt.Printf("%s v%s\n", AppName, Version)
	fmt.Printf("Go 版本: go1.21+\n")
	fmt.Println("使用 pquerna/otp 库生成兼容 Google Authenticator 的 TOTP 代码")
}
