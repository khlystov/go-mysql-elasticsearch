package river

import (
	"time"

	"github.com/juju/errors"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql-elasticsearch/elastic"
)

// Run syncs the data from MySQL and inserts to ES, skip binlog.
func (r *River) Dump() error {
	r.wg.Add(1)
	canalSyncState.Set(float64(1))
	go r.dumpLoop()

	if err := r.canal.Dump(); err != nil {
		log.Errorf("canal dump err %v", err)
		canalSyncState.Set(0)
		return errors.Trace(err)
	}

	return nil
}

func (r *River) dumpLoop() {
	bulkSize := r.c.BulkSize
	if bulkSize == 0 {
		bulkSize = 128
	}

	interval := r.c.FlushBulkTime.Duration
	if interval == 0 {
		interval = 200 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	dumping := false
	ticksAfter := 0
	defer ticker.Stop()
	defer r.wg.Done()

	reqs := make([]*elastic.BulkRequest, 0, 1024)

	for {
		needFlush := false

		select {
		case v := <-r.syncCh:
			switch v := v.(type) {
			case []*elastic.BulkRequest:
				reqs = append(reqs, v...)
				needFlush = len(reqs) >= bulkSize

				if dumping == false {
					dumping = true
					log.Info("Dumping in progress")
				}

				if ticksAfter > 0 {
					ticksAfter = 0
				}
			}
		case <-ticker.C:
			needFlush = true
			if dumping == true {
				log.Info("Chan is empty. Waiting...")
				ticksAfter++
			}
			if ticksAfter > 100 {
				log.Info("Done.")
				r.cancel()
			}
		case <-r.ctx.Done():
			return
		}

		if needFlush {
			// TODO: retry some times?
			if err := r.doBulk(reqs); err != nil {
				log.Errorf("do ES bulk err %v, close sync", err)
				r.cancel()
				return
			}
			reqs = reqs[0:0]
		}
	}
}
