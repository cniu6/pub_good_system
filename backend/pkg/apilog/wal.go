package apilog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	walFileName       = "api-access-logs.wal"
	walFilePermission = 0o600
	walDirPermission  = 0o700
)

// WAL 是访问日志的本地追加持久化队列。
// 数据库暂时不可写时，worker 会先把完整日志 fsync 到此文件；下次启动或定期重放后再删除。
// 按产品要求，WAL 保留访问日志的原始采集内容，目录必须只授予服务账号读写权限。
type WAL struct {
	dir  string
	path string
	mu   sync.Mutex
}

func NewWAL(dir string) (*WAL, error) {
	if dir == "" {
		return nil, fmt.Errorf("api access log WAL directory is empty")
	}
	if err := os.MkdirAll(dir, walDirPermission); err != nil {
		return nil, err
	}
	return &WAL{dir: dir, path: filepath.Join(dir, walFileName)}, nil
}

// Append 在单条日志完整写入磁盘并同步后才返回成功。
func (w *WAL) Append(item *models.APIAccessLog) error {
	if item == nil {
		return nil
	}
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, walFilePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}

// Replay 轮转当前写入文件后逐批重放。失败时保留完整轮转文件：
// 下次会利用 request_id 幂等写入继续重放，因此进程崩溃或中途失败不会丢日志。
func (w *WAL) Replay(batchSize int, handle func([]*models.APIAccessLog) error) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	if handle == nil {
		return nil
	}

	files, err := w.rotateAndListReplayFiles()
	if err != nil {
		return err
	}
	for _, path := range files {
		if err := replayWALFile(path, batchSize, handle); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func (w *WAL) rotateAndListReplayFiles() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if info, err := os.Stat(w.path); err == nil && info.Size() > 0 {
		replayPath := fmt.Sprintf("%s.replay.%d", w.path, time.Now().UnixNano())
		if err := os.Rename(w.path, replayPath); err != nil {
			return nil, err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) == ".tmp" {
			continue
		}
		if len(entry.Name()) > len(walFileName) && entry.Name()[:len(walFileName)] == walFileName &&
			entry.Name() != walFileName {
			files = append(files, filepath.Join(w.dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func replayWALFile(path string, batchSize int, handle func([]*models.APIAccessLog) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 单条日志可能包含 64KB 的 Body；给 JSON 外层与请求头预留空间。
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	batch := make([]*models.APIAccessLog, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := handle(batch); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for scanner.Scan() {
		var item models.APIAccessLog
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return fmt.Errorf("parse WAL entry %s: %w", path, err)
		}
		batch = append(batch, &item)
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}
