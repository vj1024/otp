# TOTP Generator

一个基于 Go 语言的双因素认证（TOTP）代码生成器，兼容 Google Authenticator。

## 功能特性

- 生成标准 6 位 TOTP 验证码（每 30 秒刷新）
- 支持从 YAML 配置文件读取密钥
- 自动刷新模式，持续获取最新验证码
- 详细模式显示剩余时间进度条
- 生成二维码 URI，方便添加到验证器 App
- 命令行友好，简洁输出便于脚本集成

## 安装

```bash
go install github.com/vj1024/otp@latest
```

或从源码构建：

```bash
git clone https://github.com/vj1024/otp.git
cd otp
go build -o otp main.go
```

## 使用方法

### 基本用法

```bash
otp -s JBSWY3DPEHPK3PXP
```

### 命令行选项

| 选项 | 简写 | 说明 |
|------|------|------|
| `--secret` | `-s` | TOTP 密钥（base32 编码）**[必需]** |
| `--config` | `-c` | 密钥配置文件路径（默认：`~/.config/otp/otp.yaml`）|
| `--verbose` | `-v` | 详细模式，显示剩余时间和进度条 |
| `--interval` | `-i` | 自动刷新间隔（秒），0 表示不刷新 |
| `--qrcode` | `-q` | 显示二维码 URI（用于添加到验证器）|
| `--issuer` | | 发行者名称（默认：`TOTP Generator`）|
| `--account` | `-a` | 账户名称（默认：`user@example.com`）|
| `--version` | | 显示版本信息 |
| `--help` | `-h` | 显示帮助信息 |

### 使用示例

#### 1. 简洁模式（默认）

```bash
$ otp -s JBSWY3DPEHPK3PXP
123456
```

#### 2. 详细模式

```bash
$ otp -s JBSWY3DPEHPK3PXP -v
==================================================
📱 TOTP Generator
--------------------------------------------------
🔢 当前代码: 123456
⏱️  当前时间: 2026-03-30 14:23:45 CST
⏳ 剩余时间: 15 秒
📊 进度: [███████████░░░░░░░░░░░]
==================================================
```

#### 3. 自动刷新

```bash
$ otp -s JBSWY3DPEHPK3PXP -i 5 -v
```

每 5 秒自动刷新显示当前验证码，按 `Ctrl+C` 退出。

#### 4. 从配置文件读取密钥

创建配置文件 `~/.config/otp/otp.yaml`：

```yaml
github: JBSWY3DPEHPK3PXP
google: KRSXG5DSN5XW4===
```

使用时：

```bash
$ otp -s github
123456
```

#### 5. 生成二维码 URI

```bash
$ otp -s JBSWY3DPEHPK3PXP -q --issuer "MyService" --account "user@example.com"
```

生成的 URI 可用于添加到 Google Authenticator 等验证器 App。

## 配置文件格式

配置文件使用 YAML 格式：

```yaml
# ~/.config/otp/otp.yaml
github: JBSWY3DPEHPK3PXP
google: KRSXG5DSN5XW4===
aws: DFMNJG3Y2F===
```

## 如何获取 Secret 密钥

1. 在大多数网站启用双因素认证时，会显示一个密钥或二维码
2. 使用二维码扫描工具或网站提供的选项获取密钥
3. 密钥通常是 Base32 编码的字符串

## 依赖库

- [go-flags](https://github.com/jessevdk/go-flags) - 命令行参数解析
- [otp](https://github.com/pquerna/otp) - TOTP/HOTP 生成库
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 解析

## 许可证

MIT License - 详见 [LICENSE](LICENSE)
