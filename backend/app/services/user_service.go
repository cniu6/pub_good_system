package services

import (
	"errors"
	"fst/backend/app/models"
	"fst/backend/internal/db"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserService 用户服务
type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// UserListQuery 用户列表查询参数
type UserListQuery struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Keyword  string `form:"keyword" json:"keyword"`
	Status   *uint8 `form:"status" json:"status"`
	GroupID  uint64 `form:"group_id" json:"group_id"`
}

// UserListResult 用户列表返回结果
type UserListResult struct {
	List     []models.User `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// GetList 分页获取用户列表
func (s *UserService) GetList(query *UserListQuery) (*UserListResult, error) {
	var users []models.User
	var total int64

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// 构建查询条件
	where := "WHERE delete_time IS NULL"
	args := []interface{}{}

	if query.Keyword != "" {
		where += " AND (username LIKE ? OR nickname LIKE ? OR email LIKE ? OR mobile LIKE ?)"
		kw := "%" + query.Keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
if query.Status != nil {
		where += " AND status = ?"
		args = append(args, *query.Status)
	}
	if query.GroupID > 0 {
		where += " AND group_id = ?"
		args = append(args, query.GroupID)
	}

	// 查询总数
	count_query := "SELECT COUNT(*) FROM users " + where
	err := db.DB.Get(&total, count_query, args...)
	if err != nil {
		return nil, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	list_query := "SELECT * FROM users " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)

	err = db.DB.Select(&users, list_query, args...)
	if err != nil {
		return nil, err
	}

	return &UserListResult{
		List:     users,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id uint64) (*models.User, error) {
	return models.GetUserByID(id)
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return models.GetUserByUsername(username)
}

// GetByEmail 根据邮箱获取用户
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	return models.GetUserByEmail(email)
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Role     string `json:"role"`
	Status   uint8  `json:"status"`
	GroupID  uint64 `json:"group_id"`
}

// Create 创建用户
func (s *UserService) Create(req *UserCreateRequest) (*models.User, error) {
	// 检查用户名是否已存在
	existing, _ := models.GetUserByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existing, _ = models.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("邮箱已存在")
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Mobile:   req.Mobile,
		Role:     req.Role,
		Status:   req.Status,
		GroupId:  req.GroupID,
		Password: req.Password, // 调用方需要先加密
	}

	if user.Role == "" {
		user.Role = "user"
	}

	err := models.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	ID       uint64 `json:"id" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Avatar   string `json:"avatar"`
	Gender   uint8  `json:"gender"`
	Birthday *int64 `json:"birthday"`
	Motto    string `json:"motto"`
	Role     string `json:"role"`
	Status   uint8  `json:"status"`
	GroupID  uint64 `json:"group_id"`
}

// Update 更新用户
func (s *UserService) Update(req *UserUpdateRequest) error {
	user, err := models.GetUserByID(req.ID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		existing, _ := models.GetUserByEmail(req.Email)
		if existing != nil && existing.ID != user.ID {
			return errors.New("邮箱已被使用")
		}
		user.Email = req.Email
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Gender > 0 {
		user.Gender = req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.Motto != "" {
		user.Motto = req.Motto
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status > 0 {
		user.Status = req.Status
	}
	if req.GroupID > 0 {
		user.GroupId = req.GroupID
	}

	now := time.Now().Unix()
	user.UpdateTime = &now

	query := `UPDATE users SET nickname = :nickname, email = :email, mobile = :mobile,
			  avatar = :avatar, gender = :gender, birthday = :birthday, motto = :motto,
			  role = :role, status = :status, group_id = :group_id, update_time = :update_time
			  WHERE id = :id`
	_, err = db.DB.NamedExec(query, user)
	return err
}

// UpdateStatus 更新用户状态
func (s *UserService) UpdateStatus(user_id uint64, status uint8) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET status = ?, update_time = ? WHERE id = ?", status, now, user_id)
	return err
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(user_id uint64, hashed_password string) error {
	return models.UpdatePassword(user_id, hashed_password)
}

// Delete 软删除用户
func (s *UserService) Delete(user_id uint64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET delete_time = ? WHERE id = ?", now, user_id)
	return err
}

// BatchDelete 批量软删除用户
func (s *UserService) BatchDelete(user_ids []uint64) error {
	if len(user_ids) == 0 {
		return nil
	}

	now := time.Now().Unix()
	query := "UPDATE users SET delete_time = ? WHERE id IN (?)"
	query, args, err := sqlx.In(query, now, user_ids)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec(query, args...)
	return err
}

// BatchUpdateStatus 批量更新用户状态
func (s *UserService) BatchUpdateStatus(user_ids []uint64, status uint8) error {
	if len(user_ids) == 0 {
		return nil
	}

	now := time.Now().Unix()
	query := "UPDATE users SET status = ?, update_time = ? WHERE id IN (?)"
	query, args, err := sqlx.In(query, status, now, user_ids)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec(query, args...)
	return err
}

// UpdateLoginInfo 更新登录信息
func (s *UserService) UpdateLoginInfo(user_id uint64, ip string) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET last_login_time = ?, last_login_ip = ?, login_failure = 0, update_time = ? WHERE id = ?",
		now, ip, now, user_id)
	return err
}

// IncrementLoginFailure 增加登录失败次数
func (s *UserService) IncrementLoginFailure(user_id uint64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET login_failure = login_failure + 1, update_time = ? WHERE id = ?", now, user_id)
	return err
}
