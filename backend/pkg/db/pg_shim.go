package db

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jackc/pgx/v5/stdlib"
)

// ========================================
// PostgreSQL 驱动兼容 shim
//
// 背景：项目里绝大多数 SQL 是直接 tx.Exec("... ? ...", args...) / db.DB.Exec(...) 写的，
// 并不会都经过 db.Q() / db.Exec() 包装（那两个是函数级的选择性包装，很多 UPDATE/INSERT
// 因为不含 MySQL 专属语法，历史上从来没被包过）。但 PostgreSQL 的驱动只认 $1,$2,... 占位符，
// 且不认反引号标识符——这两条几乎每一条 SQL 都会踩到，不可能指望每个调用点都手动改。
//
// 所以在这里注册一个包一层的 database/sql 驱动："pg-shim"：所有 SQL 文本在真正发给
// pgx 驱动之前，统一过一遍 RebindPostgresPlaceholders（?→$N）+ AdaptMySQLQueryToPostgres
// （反引号/MySQL专属函数/ON DUPLICATE KEY 等）。这样业务代码不用为了兼容 Postgres 改一行，
// 和现有 SQLite 兼容层（只改 db.Exec/db.Q 两个出口）比，这是应对"调用点太分散、不可能
// 全部收编"这个现实约束的必要选择。
// ========================================

const pgShimDriverName = "pg-shim"

func init() {
	sql.Register(pgShimDriverName, &pgShimDriver{real: stdlib.GetDefaultDriver()})
}

// adaptForPostgresWire 驱动层统一入口：先做函数/语法转换，再做占位符转换。
// 顺序不能反：ON DUPLICATE KEY 等转换里可能重排列名，与占位符无关，但保持单一入口方便追踪。
func adaptForPostgresWire(query string) string {
	return RebindPostgresPlaceholders(AdaptMySQLQueryToPostgres(query))
}

// pgShimDriver 包一层 pgx stdlib 驱动，Open 时返回包了转换逻辑的连接。
type pgShimDriver struct {
	real driver.Driver
}

func (d *pgShimDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.real.Open(name)
	if err != nil {
		return nil, err
	}
	return &pgShimConn{real: conn}, nil
}

// pgShimConn 包一层真实连接：所有携带 SQL 文本的方法转换后再转发；其它可选接口
// （Ping/ResetSession/Validator 等）直接透传，保证连接池/健康检查行为不受影响。
type pgShimConn struct {
	real driver.Conn
}

func (c *pgShimConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.real.Prepare(adaptForPostgresWire(query))
	if err != nil {
		return nil, err
	}
	return &pgShimStmt{real: stmt}, nil
}

func (c *pgShimConn) Close() error { return c.real.Close() }

func (c *pgShimConn) Begin() (driver.Tx, error) { //nolint:staticcheck // driver.Conn 老接口要求保留
	return c.real.Begin() //nolint:staticcheck
}

func (c *pgShimConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if b, ok := c.real.(driver.ConnBeginTx); ok {
		return b.BeginTx(ctx, opts)
	}
	return c.Begin()
}

func (c *pgShimConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if p, ok := c.real.(driver.ConnPrepareContext); ok {
		stmt, err := p.PrepareContext(ctx, adaptForPostgresWire(query))
		if err != nil {
			return nil, err
		}
		return &pgShimStmt{real: stmt}, nil
	}
	return c.Prepare(query)
}

func (c *pgShimConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ex, ok := c.real.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return ex.ExecContext(ctx, adaptForPostgresWire(query), args)
}

func (c *pgShimConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	qr, ok := c.real.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return qr.QueryContext(ctx, adaptForPostgresWire(query), args)
}

func (c *pgShimConn) Ping(ctx context.Context) error {
	if p, ok := c.real.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

func (c *pgShimConn) ResetSession(ctx context.Context) error {
	if r, ok := c.real.(driver.SessionResetter); ok {
		return r.ResetSession(ctx)
	}
	return nil
}

func (c *pgShimConn) IsValid() bool {
	if v, ok := c.real.(driver.Validator); ok {
		return v.IsValid()
	}
	return true
}

// pgShimStmt 包一层已 Prepare 好的语句；语句文本在 Prepare 阶段已经转换过，这里只是纯转发。
type pgShimStmt struct {
	real driver.Stmt
}

func (s *pgShimStmt) Close() error  { return s.real.Close() }
func (s *pgShimStmt) NumInput() int { return s.real.NumInput() }

func (s *pgShimStmt) Exec(args []driver.Value) (driver.Result, error) { //nolint:staticcheck
	return s.real.Exec(args) //nolint:staticcheck
}

func (s *pgShimStmt) Query(args []driver.Value) (driver.Rows, error) { //nolint:staticcheck
	return s.real.Query(args) //nolint:staticcheck
}

func (s *pgShimStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if e, ok := s.real.(driver.StmtExecContext); ok {
		return e.ExecContext(ctx, args)
	}
	return nil, driver.ErrSkip
}

func (s *pgShimStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if q, ok := s.real.(driver.StmtQueryContext); ok {
		return q.QueryContext(ctx, args)
	}
	return nil, driver.ErrSkip
}
