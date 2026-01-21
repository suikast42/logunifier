package lokishipper

//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"sync"
//	"time"
//)
//
//// LokiPushRequest represents the payload structure for Loki's push API.
//type LokiPushRequest struct {
//	Streams []LokiStream `json:"streams"`
//}
//
//type LokiStream struct {
//	Stream map[string]string `json:"stream"`
//	Values [][]string        `json:"values"`
//}
//
//// LokiBatcher handles the accumulation and periodic flushing of logs.
//type LokiBatcher struct {
//	url           string
//	labels        map[string]string
//	batchSize     int
//	flushInterval time.Duration
//
//	mu          sync.Mutex
//	pendingLogs [][]string
//	httpClient  *http.Client
//}
//
//// NewLokiBatcher initializes a batcher with a flush interval and max batch size.
//func NewLokiBatcher(url string, labels map[string]string, batchSize int, interval time.Duration) *LokiBatcher {
//	b := &LokiBatcher{
//		url:           url,
//		labels:        labels,
//		batchSize:     batchSize,
//		flushInterval: interval,
//		httpClient:    &http.Client{Timeout: 10 * time.Second},
//	}
//
//	// Start the background flusher
//	go b.runFlusher()
//	return b
//}
//
//// Log adds a log line to the current batch.
//func (b *LokiBatcher) Log(message string) {
//	b.mu.Lock()
//	defer b.mu.Unlock()
//
//	ts := fmt.Sprintf("%d", time.Now().UnixNano())
//	b.pendingLogs = append(b.pendingLogs, []string{ts, message})
//
//	// Immediate flush if we hit the batch size limit
//	if len(b.pendingLogs) >= b.batchSize {
//		go b.Flush()
//	}
//}
//
//// Flush sends the current batch to Loki.
//func (b *LokiBatcher) Flush() {
//	b.mu.Lock()
//	if len(b.pendingLogs) == 0 {
//		b.mu.Unlock()
//		return
//	}
//
//	// Copy logs and clear buffer
//	logsToSend := b.pendingLogs
//	b.pendingLogs = nil
//	b.mu.Unlock()
//
//	payload := LokiPushRequest{
//		Streams: []LokiStream{
//			{
//				Stream: b.labels,
//				Values: logsToSend,
//			},
//		},
//	}
//
//	if err := b.send(payload); err != nil {
//		fmt.Printf("[Loki Error] %v\n", err)
//	}
//}
//
//func (b *LokiBatcher) runFlusher() {
//	ticker := time.NewTicker(b.flushInterval)
//	for range ticker.C {
//		b.Flush()
//	}
//}
//
//func (b *LokiBatcher) send(payload LokiPushRequest) error {
//	body, err := json.Marshal(payload)
//	if err != nil {
//		return err
//	}
//
//	req, err := http.NewRequest("POST", b.url, bytes.NewBuffer(body))
//	if err != nil {
//		return err
//	}
//	req.Header.Set("Content-Type", "application/json")
//
//	resp, err := b.httpClient.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
//		return fmt.Errorf("loki error: %s", resp.Status)
//	}
//	return nil
//}
//
//func main() {
//	// Initialize the batcher: Flush every 5 seconds or every 100 logs.
//	batcher := NewLokiBatcher(
//		"http://localhost:3100/loki/api/v1/push",
//		map[string]string{"app": "worker-service", "env": "prod"},
//		100,
//		5*time.Second,
//	)
//
//	// Simulate high-frequency logging
//	for i := 0; i < 250; i++ {
//		batcher.Log(fmt.Sprintf("Log entry number %d", i))
//		time.Sleep(10 * time.Millisecond)
//	}
//
//	// Final flush before exiting
//	batcher.Flush()
//	fmt.Println("Batching complete.")
//}
