package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/shutter-network/nethermind-tests/config"
)

type RegisterIdentityRequest struct {
	DecryptionTimestamp int64  `json:"decryptionTimestamp"`
	IdentityPrefix      string `json:"identityPrefix"`
}

type RegisterIdentityResponse struct {
	Message RegisterIdentityMessage `json:"message"`
}

type RegisterIdentityMessage struct {
	Eon            int64  `json:"eon"`
	Identity       string `json:"identity"`
	IdentityPrefix string `json:"identity_prefix"`
	EonKey         string `json:"eon_key"`
	TxHash         string `json:"tx_hash"`
}

type DecryptionStats struct {
	totalDecryptions   atomic.Int64
	validDecryptions   atomic.Int64
	invalidDecryptions atomic.Int64
}

var (
	stats    DecryptionStats
	stopChan = make(chan os.Signal, 1)
)

type GetDataForEncryptionResponse struct {
	Eon            uint64 `json:"eon"`
	Identity       string `json:"identity"`
	IdentityPrefix string `json:"identity_prefix"`
	EonKey         string `json:"eon_key"`
	EpochID        string `json:"epoch_id"`
}

type ErrorResponse struct {
	Description string `json:"description,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	StatusCode  int    `json:"statusCode"`
}

func RunDecryptionMonitor(cfg config.Config) {
	baseURL := cfg.ShutterAPI
	interval := cfg.ApiRequestInterval
	address := cfg.ShutterRegistryCaller

	log.Printf("Starting performance monitoring\n")
	log.Printf("Base URL: %s\n", baseURL)
	log.Printf("Interval: %v\n\n", interval)

	// Setup graceful shutdown
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start monitoring in separate goroutine
	go func() {
		runFlow(baseURL, address, &wg, cfg)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runFlow(baseURL, address, &wg, cfg)
			case <-stopChan:
				return
			}
		}
	}()

	// Wait for interrupt signal
	<-stopChan
	log.Printf("Shutting down...\n")

	// Wait for all decryption requests to complete
	wg.Wait()

	// Print final statistics
	printStatistics()
}

func printStatistics() {
	log.Printf("\n=== Final Decryption Statistics ===\n")
	log.Printf("Total decryption attempts: %d\n", stats.totalDecryptions.Load())
	log.Printf("Successful decryptions: %d\n", stats.validDecryptions.Load())
	log.Printf("Failed decryptions: %d\n", stats.invalidDecryptions.Load())

	total := stats.totalDecryptions.Load()
	if total > 0 {
		successRate := float64(stats.validDecryptions.Load()) / float64(total) * 100
		log.Printf("Success rate: %.2f%%\n", successRate)
	}
}

func runFlow(baseURL, address string, wg *sync.WaitGroup, cfg config.Config) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("\n=== Performance Check at %s ===\n", timestamp)

	encryptionData, err := getDataForEncryption(baseURL, address, "")
	if err != nil {
		log.Fatalf("error in get data for encryption endpoint %s", err)
		return
	}
	decryptionTimestamp := time.Now().Unix() + 10
	registerReq := RegisterIdentityRequest{
		DecryptionTimestamp: decryptionTimestamp,
		IdentityPrefix:      encryptionData["message"].IdentityPrefix,
	}

	identity, err := registerIdentity(baseURL, registerReq)
	if err != nil {
		log.Fatalf("error encountered while registering identity %s", err)
		return
	}

	// Launch decryption key request in separate goroutine
	wg.Add(1)
	go func(identity string) {
		defer wg.Done()
		time.Sleep(cfg.DecryptionKeyWaitInterval)
		fmt.Printf("Requesting decryption key for identity: %s\n, at time: %s\n", identity, time.Now().Format("2006-01-02 15:04:05"))
		stats.totalDecryptions.Add(1)
		err = getDecryptionKey(baseURL, identity)
		if err != nil {
			log.Fatalf("error encountered while getting decryption key%s", err)
			stats.invalidDecryptions.Add(1)
		} else {
			fmt.Printf("Decryption key retrieved successfully for identity: %s\n, at time: %s\n", identity, time.Now().Format("2006-01-02 15:04:05"))
			stats.validDecryptions.Add(1)
		}
	}(identity)
}

func getDataForEncryption(baseURL, address, identityPrefix string) (map[string]GetDataForEncryptionResponse, error) {
	params := url.Values{}
	params.Add("address", address)
	if identityPrefix != "" {
		params.Add("identityPrefix", identityPrefix)
	}

	url := fmt.Sprintf("%s/get_data_for_encryption?%s", baseURL, params.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("failed to parse error response: %v", err)
		}
		return nil, err

	}

	var response map[string]GetDataForEncryptionResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse  response: %v", err)
	}

	return response, nil
}

func registerIdentity(baseURL string, req RegisterIdentityRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(
		baseURL+"/register_identity",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", fmt.Errorf("failed to parse error response: %v", err)
		}
		return "", fmt.Errorf("response error %v", errorResp.Description)
	}
	var response RegisterIdentityResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}
	fmt.Printf("Identity registered successfully: %s | tx: %s\n", response.Message.Identity, response.Message.TxHash)
	return response.Message.Identity, nil
}

func getDecryptionKey(baseURL, identity string) error {
	params := url.Values{}
	params.Add("identity", identity)

	url := fmt.Sprintf("%s/get_decryption_key?%s", baseURL, params.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("failed to parse error response, %v", err)
		}
		return fmt.Errorf("failed to parse error response, %s", errorResp.Description)
	}
	return nil
}
