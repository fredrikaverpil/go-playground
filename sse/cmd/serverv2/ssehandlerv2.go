package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type CPUInfo struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
}

func sendSSE(w http.ResponseWriter, eventName string, data any) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", eventName); err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		return err
	}

	rc := http.NewResponseController(w)
	return rc.Flush()
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	memT := time.NewTicker(time.Second)
	defer memT.Stop()

	cpuT := time.NewTicker(time.Second)
	defer cpuT.Stop()

	clientGone := r.Context().Done()

	for {
		select {
		case <-clientGone:
			log.Println("client disconnected")
			return // Exit the handler when client disconnects

		case <-memT.C:
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("error getting memory info: %s\n", err)
				continue
			}

			memInfo := MemoryInfo{
				Total:       m.Total,
				Free:        m.Free,
				Available:   m.Available,
				Used:        m.Used,
				UsedPercent: m.UsedPercent,
			}

			if err := sendSSE(w, "mem", memInfo); err != nil {
				log.Printf("error writing memory info: %s\n", err)
			}

		case <-cpuT.C:
			c, err := cpu.Times(false)
			if err != nil {
				log.Printf("error getting CPU info: %s\n", err)
				continue
			}

			cpuInfo := CPUInfo{
				User:   c[0].User,
				System: c[0].System,
				Idle:   c[0].Idle,
			}

			if err := sendSSE(w, "cpu", cpuInfo); err != nil {
				log.Printf("error writing CPU info: %s\n", err)
			}
		}
	}
}
