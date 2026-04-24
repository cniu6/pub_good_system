package main

import (
	"errors"
	"flag"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type options struct {
	Username      string
	Email         string
	Password      string
	PasswordStdin bool
	Nickname      string
	Mobile        string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "注册失败: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}

	config.InitConfig()
	db.InitDB()

	if !db.CheckTableExists("users") {
		return errors.New("users 表不存在，请先完成后端初始化")
	}

	// 检查用户名是否已存在
	existing, _ := models.GetUserByUsername(opts.Username)
	if existing != nil {
		return fmt.Errorf("用户名已存在: %s", opts.Username)
	}

	// 检查邮箱是否已存在
	existing, _ = models.GetUserByEmail(opts.Email)
	if existing != nil {
		return fmt.Errorf("邮箱已存在: %s", opts.Email)
	}

	// 加密密码（内含强度校验）
	hashed, err := utils.HashPassword(opts.Password)
	if err != nil {
		if utils.IsPasswordValidationError(err) {
			return fmt.Errorf("密码不符合要求: %v", err)
		}
		return fmt.Errorf("密码加密失败: %v", err)
	}

	nickname := opts.Nickname
	if nickname == "" {
		nickname = opts.Username
	}

	now := time.Now().Unix()
	user := &models.User{
		Username:   opts.Username,
		Email:      opts.Email,
		Nickname:   nickname,
		Mobile:     opts.Mobile,
		Password:   hashed,
		Role:       utils.AdminAuthGuard, // 管理员角色
		Status:     1,                    // 启用
		Language:   "zh-CN",
		JoinTime:   &now,
		CreateTime: &now,
		UpdateTime: &now,
	}

	if err := models.CreateUser(user); err != nil {
		return fmt.Errorf("创建管理员失败: %v", err)
	}

	fmt.Println("管理员注册成功")
	fmt.Printf("user_id: %d\n", user.ID)
	fmt.Printf("username: %s\n", user.Username)
	fmt.Printf("email: %s\n", user.Email)
	fmt.Printf("role: %s\n", user.Role)
	return nil
}

func parseOptions(args []string) (*options, error) {
	fs := flag.NewFlagSet("register_admin", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	opts := &options{}
	fs.StringVar(&opts.Username, "username", "", "管理员用户名（必填）")
	fs.StringVar(&opts.Email, "email", "", "管理员邮箱（必填）")
	fs.StringVar(&opts.Password, "password", "", "登录密码")
	fs.BoolVar(&opts.PasswordStdin, "password-stdin", false, "从标准输入读取密码")
	fs.StringVar(&opts.Nickname, "nickname", "", "昵称（默认同用户名）")
	fs.StringVar(&opts.Mobile, "mobile", "", "手机号（可选）")
	fs.Usage = func() {
		exeName := filepath.Base(os.Args[0])
		fmt.Fprintf(fs.Output(), "用法:\n")
		fmt.Fprintf(fs.Output(), "  go run ./tools/register_admin -username admin -email admin@example.com -password \"Admin123\"\n")
		fmt.Fprintf(fs.Output(), "  echo Admin123 | go run ./tools/register_admin -username admin -email admin@example.com -password-stdin\n")
		fmt.Fprintf(fs.Output(), "  %s -username admin2 -email admin2@example.com -password \"Admin123\" -nickname 管理员2\n\n", exeName)
		fmt.Fprintf(fs.Output(), "参数:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() > 0 {
		return nil, fmt.Errorf("存在未识别参数: %s", strings.Join(fs.Args(), " "))
	}

	opts.Username = strings.TrimSpace(opts.Username)
	opts.Email = strings.TrimSpace(opts.Email)
	opts.Mobile = strings.TrimSpace(opts.Mobile)
	opts.Nickname = strings.TrimSpace(opts.Nickname)

	if opts.Username == "" {
		return nil, errors.New("必须提供 -username")
	}
	if opts.Email == "" {
		return nil, errors.New("必须提供 -email")
	}

	if opts.Password != "" && opts.PasswordStdin {
		return nil, errors.New("-password 和 -password-stdin 不能同时使用")
	}
	if opts.PasswordStdin {
		password, err := readPasswordFromStdin()
		if err != nil {
			return nil, err
		}
		opts.Password = password
	}
	if opts.Password == "" {
		return nil, errors.New("请通过 -password 或 -password-stdin 提供密码")
	}

	return opts, nil
}

func readPasswordFromStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	password := strings.TrimRight(string(data), "\r\n")
	if password == "" {
		return "", errors.New("从标准输入读取到的密码为空")
	}
	return password, nil
}
