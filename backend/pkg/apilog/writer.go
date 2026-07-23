package apilog

import (
	"context"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/pkg/panicsafe"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultQueueCapacity       = 5000
	defaultQueueMaxBytes int64 = 32 << 20
	defaultBatchSize           = 100
	defaultFlushInterval       = 2 * time.Second
	walReplayInterval          = 30 * time.Second
)

// Options 定义单个应用实例的访问日志写入队列参数。
type Options struct {
	QueueCapacity int
	QueueMaxBytes int64
	BatchSize     int
	FlushInterval time.Duration
	WALDir        string
}

// Writer 负责把 HTTP 请求线程产生的日志平稳转交给单个数据库写入 worker。
// 单 worker 是刻意设计：当前小集群的单实例内不再出现多个日志 INSERT 同时争抢同一张表。
type Writer struct {
	options Options
	wal     *WAL

	queue       chan queuedEntry
	queuedBytes atomic.Int64
	accepting   atomic.Bool
	enqueueMu   sync.RWMutex

	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once

	cleanupNextAt atomic.Int64
}

type queuedEntry struct {
	item  *models.APIAccessLog
	bytes int64
}

var (
	globalWriterMu sync.RWMutex
	globalWriter   *Writer
)

// Start 根据根 .env 配置启动访问日志 writer；必须在数据库与系统设置初始化之后调用。
func Start() error {
	cfg := config.CloneGlobalConfig()
	options := Options{
		QueueCapacity: defaultQueueCapacity,
		QueueMaxBytes: defaultQueueMaxBytes,
		BatchSize:     defaultBatchSize,
		FlushInterval: defaultFlushInterval,
		WALDir:        "./api-access-log-wal",
	}
	if cfg != nil {
		options.QueueCapacity = cfg.APILogQueueCapacity
		options.QueueMaxBytes = cfg.APILogQueueMaxBytes
		options.BatchSize = cfg.APILogBatchSize
		options.FlushInterval = time.Duration(cfg.APILogFlushIntervalMillis) * time.Millisecond
		options.WALDir = cfg.APILogWALDir
	}

	writer, err := NewWriter(options)
	if err != nil {
		return err
	}
	globalWriterMu.Lock()
	if globalWriter != nil {
		globalWriterMu.Unlock()
		return nil
	}
	globalWriter = writer
	globalWriterMu.Unlock()

	// 启动 HTTP 前优先恢复上次异常退出遗留的日志；恢复失败不阻断服务，
	// WAL 文件会保留，后台 worker 会在后续周期再次重试。
	if err := writer.replayWAL(); err != nil {
		log.Printf("[APILogWriter] 启动重放 WAL 失败，稍后自动重试: %v", err)
	}
	panicsafe.Go("APILogWriter.loop", writer.loop)
	return nil
}

// Enqueue 将日志投递到当前实例的有界内存队列。队列满或字节超限时同步 fsync 到 WAL，
// 因而不会用无限 goroutine 或无限数组把数据库积压转化为内存 OOM。
func Enqueue(item *models.APIAccessLog) error {
	globalWriterMu.RLock()
	writer := globalWriter
	globalWriterMu.RUnlock()
	if writer == nil {
		return errors.New("api access log writer is not started")
	}
	return writer.Enqueue(item)
}

// Stop 停止接收新日志并在 ctx 截止前刷空内存队列。
// 超时后剩余的事件由 worker 写入 WAL，下一次启动会继续重放。
func Stop(ctx context.Context) error {
	globalWriterMu.RLock()
	writer := globalWriter
	globalWriterMu.RUnlock()
	if writer == nil {
		return nil
	}
	return writer.Stop(ctx)
}

func NewWriter(options Options) (*Writer, error) {
	options = normalizeOptions(options)
	wal, err := NewWAL(options.WALDir)
	if err != nil {
		return nil, err
	}
	writer := &Writer{
		options: options,
		wal:     wal,
		queue:   make(chan queuedEntry, options.QueueCapacity),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	writer.accepting.Store(true)
	return writer, nil
}

func normalizeOptions(options Options) Options {
	if options.QueueCapacity <= 0 {
		options.QueueCapacity = defaultQueueCapacity
	}
	if options.QueueMaxBytes <= 0 {
		options.QueueMaxBytes = defaultQueueMaxBytes
	}
	if options.BatchSize <= 0 {
		options.BatchSize = defaultBatchSize
	}
	if options.FlushInterval <= 0 {
		options.FlushInterval = defaultFlushInterval
	}
	if options.WALDir == "" {
		options.WALDir = "./api-access-log-wal"
	}
	return options
}

func (w *Writer) Enqueue(item *models.APIAccessLog) error {
	if item == nil {
		return nil
	}
	w.enqueueMu.RLock()
	defer w.enqueueMu.RUnlock()
	if !w.accepting.Load() {
		return w.wal.Append(item)
	}

	size := estimateLogBytes(item)
	for {
		current := w.queuedBytes.Load()
		if current+size > w.options.QueueMaxBytes {
			return w.wal.Append(item)
		}
		if w.queuedBytes.CompareAndSwap(current, current+size) {
			break
		}
	}

	entry := queuedEntry{item: item, bytes: size}
	select {
	case w.queue <- entry:
		return nil
	default:
		// channel 已满时立即撤销预占的字节，并同步落 WAL；请求线程不等待数据库。
		w.queuedBytes.Add(-size)
		return w.wal.Append(item)
	}
}

func (w *Writer) loop() {
	defer close(w.doneCh)

	flushTicker := time.NewTicker(w.options.FlushInterval)
	replayTicker := time.NewTicker(walReplayInterval)
	defer flushTicker.Stop()
	defer replayTicker.Stop()

	batch := make([]queuedEntry, 0, w.options.BatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		items := make([]*models.APIAccessLog, 0, len(batch))
		for _, entry := range batch {
			items = append(items, entry.item)
			w.queuedBytes.Add(-entry.bytes)
		}
		if err := w.persist(items, true); err != nil {
			log.Printf("[APILogWriter] 批量写入失败，已转存 WAL: %v", err)
		}
		batch = batch[:0]
	}
	drainToWAL := func() {
		flush()
		for {
			select {
			case entry := <-w.queue:
				w.queuedBytes.Add(-entry.bytes)
				if err := w.wal.Append(entry.item); err != nil {
					log.Printf("[APILogWriter] 关闭时落 WAL 失败: %v", err)
				}
			default:
				return
			}
		}
	}

	for {
		select {
		case entry := <-w.queue:
			batch = append(batch, entry)
			if len(batch) >= w.options.BatchSize {
				flush()
			}
		case <-flushTicker.C:
			flush()
		case <-replayTicker.C:
			flush()
			if err := w.replayWAL(); err != nil {
				log.Printf("[APILogWriter] 定期重放 WAL 失败: %v", err)
			}
		case <-w.stopCh:
			drainToWAL()
			return
		}
	}
}

func (w *Writer) persist(items []*models.APIAccessLog, fallbackToWAL bool) error {
	created, err := models.CreateAPIAccessLogs(items)
	if err != nil {
		if !fallbackToWAL {
			// 当前批次本来就来自 WAL；保留轮转文件等待下次重放，不能再次追加造成无限复制。
			return err
		}
		for _, item := range items {
			if walErr := w.wal.Append(item); walErr != nil {
				return fmt.Errorf("database write failed: %w; WAL append failed: %v", err, walErr)
			}
		}
		return err
	}

	for _, item := range created {
		if err := models.RecordAPIAccessLogAggregate(item); err != nil {
			// 主日志已成功持久化；聚合仅影响统计面板，记录错误后等待后续维护回填。
			log.Printf("[APILogWriter] 汇总更新失败 request_id=%s: %v", item.RequestID, err)
		}
	}
	w.scheduleRetentionCleanup()
	return nil
}

func (w *Writer) replayWAL() error {
	return w.wal.Replay(w.options.BatchSize, func(items []*models.APIAccessLog) error {
		return w.persist(items, false)
	})
}

func (w *Writer) scheduleRetentionCleanup() {
	cfg := services.GetGlobalAPILogRuntimeConfig()
	if cfg.MaxCount <= 0 && !(cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0) {
		return
	}

	now := time.Now().UnixNano()
	nextAt := w.cleanupNextAt.Load()
	if nextAt > now {
		return
	}
	interval := time.Duration(cfg.CleanupIntervalSeconds) * time.Second
	if !w.cleanupNextAt.CompareAndSwap(nextAt, now+int64(interval)) {
		return
	}

	if cfg.MaxCount > 0 {
		if _, err := models.CleanExcessAPIAccessLogs(cfg.MaxCount); err != nil {
			w.cleanupNextAt.Store(now)
			log.Printf("[APILogWriter] 自动清理超限日志失败: %v", err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := models.CleanExcessAPIAccessLogsPerUser(cfg.PerUserMaxCount); err != nil {
			log.Printf("[APILogWriter] 按用户清理超限日志失败: %v", err)
		}
	}
}

func (w *Writer) Stop(ctx context.Context) error {
	w.stopOnce.Do(func() {
		// 先阻止所有已开始但尚未真正发送到 channel 的 Enqueue，
		// 再关闭 worker；这样 drain 完成后不会有“晚到”的内存日志丢失。
		w.enqueueMu.Lock()
		w.accepting.Store(false)
		close(w.stopCh)
		w.enqueueMu.Unlock()
	})
	select {
	case <-w.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func estimateLogBytes(item *models.APIAccessLog) int64 {
	// 估算值用于内存背压，不要求精确；额外预留结构体、指针、map 编码等运行时开销。
	size := 512 + len(item.RequestID) + len(item.Username) + len(item.Role) + len(item.AuthMethod) +
		len(item.Scene) + len(item.Method) + len(item.Transport) + len(item.Protocol) + len(item.Path) +
		len(item.RoutePath) + len(item.HandlerName) + len(item.RequestContentType) + len(item.ResponseContentType) +
		len(item.QueryString) + len(item.IP) + len(item.SourceIP) + len(item.XIP) + len(item.XForwardedFor) +
		len(item.XRealIP) + len(item.UserAgent) + len(item.Referer)
	for _, value := range []*string{item.PathParams, item.RequestHeaders, item.RequestBody, item.ResponseBody} {
		if value != nil {
			size += len(*value)
		}
	}
	return int64(size)
}

// Pending 返回当前内存队列中的条数，仅用于监控与测试。
func (w *Writer) Pending() int {
	return len(w.queue)
}

// QueueBytes 返回当前内存队列预估字节数，仅用于监控与测试。
func (w *Writer) QueueBytes() int64 {
	return w.queuedBytes.Load()
}
