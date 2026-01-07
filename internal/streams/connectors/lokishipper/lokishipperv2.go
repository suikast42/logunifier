package lokishipper

//import (
//	"io"
//	"net/http"
//	"time"
//
//	"github.com/grafana/loki/v3/pkg/logproto"
//	"github.com/prometheus/common/model"
//)
//
//package main
//
//import (
//"bytes"
//"compress/gzip"
//"fmt"
//"io"
//"net/http"
//"time"
//
//"github.com/gogo/protobuf/proto"
//"github.com/golang/snappy"
//"github.com/prometheus/common/model"
//"github.com/grafana/loki/pkg/logproto"
//)

//type LokiPusher struct {
//	url     string
//	client  *http.Client
//	labels  model.LabelSet
//}
//
//func NewLokiPusher(lokiURL string, staticLabels map[string]string) *LokiPusher {
//	labels := model.LabelSet{
//		"app":      "my-go-service",
//		"env":      "production",
//		"instance": "backend-01",
//	}
//	for k, v := range staticLabels {
//		labels[model.LabelName(k)] = model.LabelValue(v)
//	}
//
//	return &LokiPusher{
//		url:    lokiURL + "/loki/api/v1/push",
//		client: &http.Client{Timeout: 10 * time.Second},
//		labels: labels,
//	}
//}
//
//func (p *LokiPusher) PushLog(line string, extraLabels map[string]string) error {
//	// Merge static + dynamic labels
//	streamLabels := p.labels.Clone()
//	for k, v := range extraLabels {
//		streamLabels[model.LabelName(k)] = model.LabelValue(v)
//	}
//
//	// Create entry
//	entry := &logproto.Entry{
//		Timestamp: time.Now(),
//		Line:      line,
//	}
//
//	// Create stream
//	stream := logproto.Stream{
//		Labels: streamLabels.String(),
//		Entries: []logproto.Entry{*entry},
//	}
//
//	// Build push request
//	req := logproto.PushRequest{
//		Streams: []logproto.Stream{stream},
//	}
//
//	// Serialize to protobuf
//	data, err := proto.Marshal(&req)
//	if err != nil {
//		return fmt.Errorf("protobuf marshal error: %w", err)
//	}
//
//	// Compress with Snappy
//	compressed := snappy.Encode(nil, data)
//
//	// Send HTTP request
//	httpReq, err := http.NewRequest("POST", p.url, bytes.NewReader(compressed))
//	if err != nil {
//		return err
//	}
//
//	httpReq.Header.Set("Content-Type", "application/x-protobuf")
//	httpReq.Header.Set("X-Scope-OrgID", "1") // for multi-tenant, otherwise omit or set tenant ID
//
//	resp, err := p.client.Do(httpReq)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode >= 400 {
//		body, _ := io.ReadAll(resp.Body)
//		return fmt.Errorf("loki rejected logs: %d %s", resp.StatusCode, string(body))
//	}
//
//	return nil
//}
//
//// Example usage
//func main() {
//	pusher := NewLokiPusher("http://loki:3100", map[string]string{"service": "auth-api"})
//
//	for i := 0; i < 10; i++ {
//		err := pusher.PushLog(
//			fmt.Sprintf("User login successful, user_id=123%d", i),
//			map[string]string{"level": "info", "user_id": fmt.Sprintf("123%d", i)},
//		)
//		if err != nil {
//			fmt.Printf("Failed to send log: %v\n", err)
//		} else {
//			fmt.Println("Log sent to Loki")
//		}
//		time.Sleep(1 * time.Second)
//	}
//}
