package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type options struct {
	ID            uint64
	Username      string
	Email         string
	Account       string
	Password      string
	PasswordStdin bool
	ClearLock     bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "修改失败: %v\n", err)
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
	models.InitUserSessionsTable()
	if !db.CheckTableExists("user_sessions") {
		return errors.New("user_sessions 表不存在，无法安全修改密码，请先完成后端初始化")
	}

	user, err := loadTargetUser(opts)
	if err != nil {
		return err
	}

	authService := services.NewAuthService()
	if err := authService.UpdatePassword(user.ID, opts.Password); err != nil {
		return err
	}

	if opts.ClearLock {
		if err := services.NewUserService().ClearLockUntil(user.ID); err != nil {
			return err
		}
	}

	fmt.Println("密码修改成功")
	fmt.Printf("user_id: %d\n", user.ID)
	fmt.Printf("username: %s\n", user.Username)
	fmt.Printf("email: %s\n", user.Email)
	if opts.ClearLock {
		fmt.Println("已清除登录失败次数与锁定状态")
	}

	return nil
}

func parseOptions(args []string) (*options, error) {
	fs := flag.NewFlagSet("reset_password", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	opts := &options{}
	fs.Uint64Var(&opts.ID, "id", 0, "按用户ID查找")
	fs.StringVar(&opts.Username, "username", "", "按用户名查找")
	fs.StringVar(&opts.Email, "email", "", "按邮箱查找")
	fs.StringVar(&opts.Account, "account", "", "按登录账号查找（用户名或邮箱）")
	fs.StringVar(&opts.Password, "password", "", "新密码")
	fs.BoolVar(&opts.PasswordStdin, "password-stdin", false, "从标准输入读取新密码")
	fs.BoolVar(&opts.ClearLock, "clear-lock", true, "修改后同时清除登录失败次数和锁定状态")
	fs.Usage = func() {
		exeName := filepath.Base(os.Args[0])
		fmt.Fprintf(fs.Output(), "用法:\n")
		fmt.Fprintf(fs.Output(), "  go run ./tools/reset_password -username admin -password \"NewPass123\"\n")
		fmt.Fprintf(fs.Output(), "  echo NewPass123 | go run ./tools/reset_password -email admin@example.com -password-stdin\n")
		fmt.Fprintf(fs.Output(), "  %s -account admin@example.com -password \"NewPass123\"\n\n", exeName)
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
	opts.Account = strings.TrimSpace(opts.Account)

	identifierCount := 0
	if opts.ID > 0 {
		identifierCount++
	}
	if opts.Username != "" {
		identifierCount++
	}
	if opts.Email != "" {
		identifierCount++
	}
	if opts.Account != "" {
		identifierCount++
	}
	if identifierCount != 1 {
		return nil, errors.New("必须且只能提供一个用户定位参数：-id / -username / -email / -account")
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
		return nil, errors.New("请通过 -password 或 -password-stdin 提供新密码")
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

func loadTargetUser(opts *options) (*models.User, error) {
	switch {
	case opts.ID > 0:
		return getUser(func() (*models.User, error) {
			return models.GetUserByID(opts.ID)
		}, fmt.Sprintf("ID=%d", opts.ID))
	case opts.Username != "":
		return getUser(func() (*models.User, error) {
			return models.GetUserByUsername(opts.Username)
		}, fmt.Sprintf("username=%s", opts.Username))
	case opts.Email != "":
		return getUser(func() (*models.User, error) {
			return models.GetUserByEmail(opts.Email)
		}, fmt.Sprintf("email=%s", opts.Email))
	default:
		return getUser(func() (*models.User, error) {
			return models.GetUserByUsernameOrEmail(opts.Account)
		}, fmt.Sprintf("account=%s", opts.Account))
	}
}

func getUser(load func() (*models.User, error), desc string) (*models.User, error) {
	user, err := load()
	if err == nil {
		return user, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("未找到目标用户: %s", desc)
	}
	return nil, err
}
