package metric

import (
	"time"
	"sync/atomic"
	"fmt"
	"encoding/json"

	"redis-shake/base"
	"redis-shake/configure"
	"pkg/libs/log"
)

const (
	updateInterval = 10 // seconds
)

var (
	MetricVar *Metric
	runner base.Runner
)

type Op interface {
	Update()
}

type Percent struct {
	Dividend uint64
	Divisor  uint64
}

func (p *Percent) Set(dividend, divisor uint64) {
	atomic.AddUint64(&p.Dividend, dividend)
	atomic.AddUint64(&p.Divisor, divisor)
}

// input: return string?
func (p *Percent) Get(returnString bool) interface{} {
	if divisor := atomic.LoadUint64(&p.Divisor); divisor == 0 {
		if returnString {
			return "null"
		} else {
			return int64(^uint64(0) >> 1) // int64_max
		}
	} else {
		dividend := atomic.LoadUint64(&p.Dividend)
		if returnString {
			return fmt.Sprintf("%.02f", float64(dividend)/float64(divisor))
		} else {
			return dividend / divisor
		}
	}
}

func (p *Percent) Update() {
	p.Dividend = 0
	p.Divisor = 0
}

type Delta struct {
	Value uint64 // current value
}

func (d *Delta) Update() {
	d.Value = 0
}

type Combine struct {
	Total uint64 // total number
	Delta        // delta
}

func (c *Combine) Set(val uint64) {
	atomic.AddUint64(&c.Delta.Value, val)
	atomic.AddUint64(&(c.Total), val)
}

// main struct
type Metric struct {
	PullCmdCount    Combine
	BypassCmdCount  Combine
	PushCmdCount    Combine
	SuccessCmdCount Combine
	FailCmdCount    Combine

	Delay       Percent // ms
	AvgDelay    Percent // ms
	NetworkFlow Combine // +speed

	FullSyncProgress uint64
}

func CreateMetric(r base.Runner) {
	runner = r
	MetricVar = &Metric{}

	go MetricVar.run()
}

func (m *Metric) resetEverySecond(items []Op) {
	for _, item := range items {
		item.Update()
	}
}

func (m *Metric) run() {
	resetItems := []Op{
		&m.PullCmdCount.Delta,
		&m.BypassCmdCount.Delta,
		&m.PushCmdCount.Delta,
		&m.SuccessCmdCount.Delta,
		&m.FailCmdCount.Delta,
		&m.Delay,
		&m.NetworkFlow.Delta,
	}

	go func() {
		tick := 0
		for range time.NewTicker(1 * time.Second).C {
			tick++
			if tick % updateInterval == 0 && conf.Options.MetricPrintLog {
				stat := NewMetricRest()
				if opts, err := json.Marshal(stat); err != nil {
					log.Infof("marshal metric stat error[%v]", err)
				} else {
					log.Info(string(opts))
				}
			}

			m.resetEverySecond(resetItems)
		}
	}()
}

func (m *Metric) AddPullCmdCount(val uint64) {
	m.PullCmdCount.Set(val)
}

func (m *Metric) GetPullCmdCount() interface{} {
	return atomic.LoadUint64(&m.PullCmdCount.Value)
}

func (m *Metric) GetPullCmdCountTotal() interface{} {
	return atomic.LoadUint64(&m.PullCmdCount.Total)
}

func (m *Metric) AddBypassCmdCount(val uint64) {
	m.BypassCmdCount.Set(val)
}

func (m *Metric) GetBypassCmdCount() interface{} {
	return atomic.LoadUint64(&m.BypassCmdCount.Value)
}

func (m *Metric) GetBypassCmdCountTotal() interface{} {
	return atomic.LoadUint64(&m.BypassCmdCount.Total)
}

func (m *Metric) AddPushCmdCount(val uint64) {
	m.PushCmdCount.Set(val)
}

func (m *Metric) GetPushCmdCount() interface{} {
	return atomic.LoadUint64(&m.PushCmdCount.Value)
}

func (m *Metric) GetPushCmdCountTotal() interface{} {
	return atomic.LoadUint64(&m.PushCmdCount.Total)
}

func (m *Metric) AddSuccessCmdCount(val uint64) {
	m.SuccessCmdCount.Set(val)
}

func (m *Metric) GetSuccessCmdCount() interface{} {
	return atomic.LoadUint64(&m.SuccessCmdCount.Value)
}

func (m *Metric) GetSuccessCmdCountTotal() interface{} {
	return atomic.LoadUint64(&m.SuccessCmdCount.Total)
}

func (m *Metric) AddFailCmdCount(val uint64) {
	m.FailCmdCount.Set(val)
}

func (m *Metric) GetFailCmdCount() interface{} {
	return atomic.LoadUint64(&m.FailCmdCount.Value)
}

func (m *Metric) GetFailCmdCountTotal() interface{} {
	return atomic.LoadUint64(&m.FailCmdCount.Total)
}

func (m *Metric) AddDelay(val uint64) {
	m.Delay.Set(val, 1)
	m.AvgDelay.Set(val, 1)
}

func (m *Metric) GetDelay() interface{} {
	return m.Delay.Get(true)
}

func (m *Metric) GetAvgDelay() interface{} {
	return m.AvgDelay.Get(true)
}

func (m *Metric) AddNetworkFlow(val uint64) {
	// atomic.AddUint64(&m.NetworkFlow.Value, val)
	m.NetworkFlow.Set(val)
}

func (m *Metric) GetNetworkFlow() interface{} {
	return atomic.LoadUint64(&m.NetworkFlow.Value)
}

func (m *Metric) GetNetworkFlowTotal() interface{} {
	return atomic.LoadUint64(&m.NetworkFlow.Total)
}

func (m *Metric) SetFullSyncProgress(val uint64) {
	m.FullSyncProgress = val
}

func (m *Metric) GetFullSyncProgress() interface{} {
	return m.FullSyncProgress
}
