package service

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
)

type MetricBuffer struct {
	db            *gorm.DB
	buffer        chan *model.MetricRecord
	batchSize     int
	flushInterval time.Duration
	done          chan struct{}
	wg            sync.WaitGroup
}

func NewMetricBuffer(db *gorm.DB, batchSize int, flushInterval time.Duration) *MetricBuffer {
	if batchSize <= 0 {
		batchSize = 100
	}
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second
	}
	return &MetricBuffer{
		db:            db,
		buffer:        make(chan *model.MetricRecord, 10000),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
	}
}

func (mb *MetricBuffer) Submit(record *model.MetricRecord) {
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	select {
	case mb.buffer <- record:
	default:
		logger.Warnf("Metric buffer full, dropping record")
	}
}

func (mb *MetricBuffer) Start() {
	mb.wg.Add(1)
	go mb.run()
}

func (mb *MetricBuffer) Stop() {
	close(mb.done)
	mb.wg.Wait()
}

func (mb *MetricBuffer) Flush() {
	batch := mb.drain()
	if len(batch) > 0 {
		mb.writeBatch(batch)
	}
}

func (mb *MetricBuffer) run() {
	defer mb.wg.Done()

	ticker := time.NewTicker(mb.flushInterval)
	defer ticker.Stop()

	batch := make([]*model.MetricRecord, 0, mb.batchSize)

	for {
		select {
		case record := <-mb.buffer:
			batch = append(batch, record)
			if len(batch) >= mb.batchSize {
				mb.writeBatch(batch)
				batch = make([]*model.MetricRecord, 0, mb.batchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				mb.writeBatch(batch)
				batch = make([]*model.MetricRecord, 0, mb.batchSize)
			}
		case <-mb.done:
			// 排空 channel 中剩余记录
			for {
				select {
				case record := <-mb.buffer:
					batch = append(batch, record)
				default:
					if len(batch) > 0 {
						mb.writeBatch(batch)
					}
					return
				}
			}
		}
	}
}

func (mb *MetricBuffer) drain() []*model.MetricRecord {
	var batch []*model.MetricRecord
	for {
		select {
		case record := <-mb.buffer:
			batch = append(batch, record)
		default:
			return batch
		}
	}
}

func (mb *MetricBuffer) writeBatch(records []*model.MetricRecord) {
	if len(records) == 0 {
		return
	}
	if err := mb.db.CreateInBatches(records, mb.batchSize).Error; err != nil {
		logger.Errorf("Metric batch write error: %v", err)
	}
}
