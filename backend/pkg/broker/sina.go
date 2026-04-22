package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"trading/model"
)

// SinaBroker 新浪财经行情数据提供者
type SinaBroker struct {
	client     *http.Client
	baseURL    string
	maxRetries int
	retryDelay time.Duration
}

// NewSinaBroker 创建新浪财经数据提供者
func NewSinaBroker() *SinaBroker {
	return &SinaBroker{
		client:     &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://hq.sinajs.cn",
		maxRetries: 3,
		retryDelay: 500 * time.Millisecond,
	}
}

func (p *SinaBroker) getBytes(ctx context.Context, path string, headers map[string]string) ([]byte, error) {
	u, err := url.Parse(p.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(p.retryDelay * time.Duration(attempt))
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create request failed: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: status=%d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("client error: status=%d", resp.StatusCode)
		}

		return body, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", p.maxRetries, lastErr)
}

// GetStockTodayInBatch 批量获取今日行情数据
func (p *SinaBroker) GetStockTodayInBatch(ctx context.Context, codes []string) (map[string]*model.StockKline, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	body, err := p.getBytes(ctx, fmt.Sprintf("/list=%s", strings.Join(codes, ",")), map[string]string{
		"Referer": "https://finance.sina.com.cn/",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch realtime batch failed: %w", err)
	}

	utf8Body, err := gbkToUTF8(body)
	if err != nil {
		return nil, fmt.Errorf("gbk to utf8 conversion failed: %w", err)
	}

	rawMap := parseRealtimeResponse(string(utf8Body))
	result := make(map[string]*model.StockKline, len(rawMap))
	for code, content := range rawMap {
		if kline := parseRealtimeToKline(code, content); kline != nil {
			result[code] = kline
		}
	}
	return result, nil
}

// GetStockToday 获取单个代码的今日数据
func (p *SinaBroker) GetStockToday(ctx context.Context, code string) (*model.StockKline, error) {
	result, err := p.GetStockTodayInBatch(ctx, []string{code})
	if err != nil {
		return nil, err
	}
	kline, ok := result[code]
	if !ok {
		return nil, fmt.Errorf("no data for code: %s", code)
	}
	return kline, nil
}

// GetStockHistorical 获取历史 K 线数据
func (p *SinaBroker) GetStockHistorical(ctx context.Context, symbol string) ([]model.StockKline, error) {
	url := fmt.Sprintf(
		"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=240&ma=no&datalen=30",
		symbol,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status=%d", resp.StatusCode)
	}

	return parseKLineResponse(symbol, string(body))
}

// --- parser ---

var realtimeRegex = regexp.MustCompile(`var hq_str_([^=]+)="([^"]*)"`)

func parseRealtimeResponse(body string) map[string]string {
	result := make(map[string]string)
	for _, m := range realtimeRegex.FindAllStringSubmatch(body, -1) {
		if len(m) >= 3 && m[2] != "" {
			result[m[1]] = m[2]
		}
	}
	return result
}

func parseRealtimeToKline(code, content string) *model.StockKline {
	fields := strings.Split(content, ",")
	if len(fields) < 5 {
		return nil
	}

	if IsStockCode(code) && len(fields) >= 32 {
		return parseStockToKline(code, fields)
	}
	return parseExchangeToKline(code, fields)
}

func parseStockToKline(code string, fields []string) *model.StockKline {
	open, _ := strconv.ParseFloat(fields[1], 64)
	high, _ := strconv.ParseFloat(fields[4], 64)
	low, _ := strconv.ParseFloat(fields[5], 64)
	closePrice, _ := strconv.ParseFloat(fields[3], 64)
	volume, _ := strconv.ParseInt(fields[8], 10, 64)

	date := fields[30]
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	return &model.StockKline{
		Code:   code,
		Date:   date,
		Open:   open,
		High:   high,
		Low:    low,
		Close:  closePrice,
		Volume: volume,
	}
}

func parseExchangeToKline(code string, fields []string) *model.StockKline {
	open, _ := strconv.ParseFloat(fields[1], 64)
	closePrice, _ := strconv.ParseFloat(fields[2], 64)
	high, _ := strconv.ParseFloat(fields[3], 64)
	volume, _ := strconv.ParseInt(fields[4], 10, 64)

	var date string
	if len(fields) > 10 {
		date = fields[10]
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	return &model.StockKline{
		Code:   code,
		Date:   date,
		Open:   open,
		High:   high,
		Low:    open,
		Close:  closePrice,
		Volume: volume,
	}
}

// IsStockCode 判断是否为股票代码（sh/sz/hk 开头）
func IsStockCode(code string) bool {
	return strings.HasPrefix(code, "sh") ||
		strings.HasPrefix(code, "sz") ||
		strings.HasPrefix(code, "hk")
}

// --- helpers ---

func gbkToUTF8(data []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	return io.ReadAll(reader)
}

type kLineItem struct {
	Day    string `json:"day"`
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
}

func parseKLineResponse(symbol, body string) ([]model.StockKline, error) {
	var items []kLineItem
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		return nil, fmt.Errorf("parse kline failed: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no kline data")
	}

	result := make([]model.StockKline, 0, len(items))
	for _, item := range items {
		open, _ := strconv.ParseFloat(item.Open, 64)
		high, _ := strconv.ParseFloat(item.High, 64)
		low, _ := strconv.ParseFloat(item.Low, 64)
		closePrice, _ := strconv.ParseFloat(item.Close, 64)
		volume, _ := strconv.ParseInt(item.Volume, 10, 64)

		result = append(result, model.StockKline{
			Code:   symbol,
			Date:   item.Day,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}
	return result, nil
}
