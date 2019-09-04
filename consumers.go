package pulsar

import (
	"context"
	"sync"
	"time"

	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tylog"
)

type ConsumerList struct {
	list             []*consumerImpl
	FlowPeriodSecond int
	FlowPermit       uint32
}

func (l *ConsumerList) ReceiveAndHandle(ctx context.Context, handler PayloadHandler) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(l.list); i++ {
		safe := i
		wg.Add(1)
		go func() {
			l.list[safe].ReceiveAndHandle(ctx, handler)
			wg.Done()
		}()
	}
	go l.CronFlow()
	wg.Wait()
}

func (l *ConsumerList) CronFlow() {
	if l.FlowPeriodSecond == 0 {
		return
	}
	if l.FlowPermit == 0 {
		return
	}
	tk := time.NewTicker(time.Duration(l.FlowPeriodSecond) * time.Second)
	for range tk.C {
		for i := 0; i < len(l.list); i++ {
			c := l.list[i].csm.Consumer(context.Background())
			if c != nil {
				tylog.Debug("cron flow", tylog.String("topic", c.Topic))
				err := c.Flow(l.FlowPermit)
				if err != nil {
					tylog.Error("flow failed", tylog.ErrorField(err), tylog.String("topic", c.Topic))
				}
			}
		}
	}
}