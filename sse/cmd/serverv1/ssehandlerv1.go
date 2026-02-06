package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

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

	rc := http.NewResponseController(w)

	for {
		select {
		case <-clientGone:
			log.Println("client disconnected")
			return

		case <-memT.C:
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("error getting memory info: %s\n", err)
			}
			if _, err := fmt.Fprintf(
				w,
				"event:mem\ndata:Total: %d\ndata:Free: %d\ndata:Available: %d\ndata:Used: %d\ndata:UsedPercent: %f\n\n",
				m.Total,
				m.Free,
				m.Available,
				m.Used,
				m.UsedPercent,
			); err != nil {
				log.Printf("error writing memory info: %s\n", err)
			}
			if err := rc.Flush(); err != nil {
				log.Printf("error flushing: %s\n", err)
			}

		case <-cpuT.C:
			c, err := cpu.Times(false)
			if err != nil {
				log.Printf("error getting memory info: %s\n", err)
			}
			if _, err := fmt.Fprintf(
				w,
				"event:cpu\ndata:User: %f\ndata:System: %f\ndata:Idle: %f\n\n",
				c[0].User,
				c[0].System,
				c[0].Idle,
			); err != nil {
				log.Printf("error writing cpu info: %s\n", err)
			}
			if err := rc.Flush(); err != nil {
				log.Printf("error flushing: %s\n", err)
			}
		}
	}
}
