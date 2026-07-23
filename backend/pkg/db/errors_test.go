package db

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsDuplicateKeyError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"普通错误", errors.New("connection refused"), false},
		{"MySQL 1062", &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}, true},
		{"MySQL 其他", &mysql.MySQLError{Number: 1045, Message: "Access denied"}, false},
		{"SQLite", errors.New("UNIQUE constraint failed: payment_orders.order_no"), true},
		{"Postgres", errors.New("ERROR: duplicate key value violates unique constraint \"idx_order_no\" (SQLSTATE 23505)"), true},
		{"MySQL 文本", errors.New("Error 1062: Duplicate entry 'x' for key 'PRIMARY'"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsDuplicateKeyError(tc.err); got != tc.want {
				t.Fatalf("IsDuplicateKeyError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsRetryableTransactionError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"普通错误", errors.New("connection refused"), false},
		{"MySQL 死锁", &mysql.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock"}, true},
		{"MySQL 锁等待超时", &mysql.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"}, true},
		{"MySQL 其他错误", &mysql.MySQLError{Number: 1045, Message: "Access denied"}, false},
		{"Postgres 死锁", errors.New("ERROR: deadlock detected (SQLSTATE 40P01)"), true},
		{"Postgres 序列化冲突", errors.New("ERROR: could not serialize access due to concurrent update (SQLSTATE 40001)"), true},
		{"SQLite 锁占用", errors.New("database is locked"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRetryableTransactionError(tc.err); got != tc.want {
				t.Fatalf("IsRetryableTransactionError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
