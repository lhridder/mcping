package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-mc/mcping"
	"github.com/mattn/go-colorable"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	PromListen string   `json:"promListen"`
	Targets    []string `json:"targets"`
}

var promListen = new(string)
var targets = new([]string)
var protocol = flag.Int("p", 578, "The minecraft protocol version")
var output = colorable.NewColorableStdout()

// Define prometheus counters
var (
	playerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mcping_playercount",
		Help: "Number of connected players",
	}, []string{"host"})
	pingDelay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mcping_pingdelay",
		Help: "Delay when dialing server",
	}, []string{"host"})
)

// Main function
func main() {
	// Fetch config
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var config Config
	jsonParser := json.NewDecoder(jsonFile)
	err = jsonParser.Decode(&config)
	jsonFile.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	*promListen = config.PromListen
	*targets = config.Targets
	// Start prom listener
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(*promListen, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	// Start fetching targets
	for {
		fmt.Println("Fetching all targets...")
		for _, host := range *targets {
			fmt.Println("Fetching " + host)
			players, delay := getServerStats(host)
			playerCount.With(prometheus.Labels{"host": host}).Set(float64(players))
			pingDelay.With(prometheus.Labels{"host": host}).Set(float64(delay))
			fmt.Println("Fetched " + host + ": " + strconv.Itoa(players))
		}
		time.Sleep(time.Minute)
	}
}

// Get target stats
func getServerStats(host string) (playercount int, delay time.Duration) {
	addrs := lookupMC(host)
	for _, addr := range addrs {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Fprintf(output, "dial error: %v\n", err)
			continue
		}
		hostname, _, err := net.SplitHostPort(addr)
		status, delay, err := mcping.PingAndListConn(conn, *protocol, hostname)
		if err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			continue
		}
		return status.Players.Online, delay
	}
	return
}

// Resolve domain and/or SRV record
func lookupMC(addr string) (addrs []string) {
	if !strings.Contains(addr, ":") {
		_, addrsSRV, err := net.LookupSRV("minecraft", "tcp", addr)
		if err == nil && len(addrsSRV) > 0 {
			for _, addrSRV := range addrsSRV {
				addrs = append(addrs, net.JoinHostPort(addrSRV.Target, strconv.Itoa(int(addrSRV.Port))))
			}
			return
		}
		return []string{net.JoinHostPort(addr, "25565")}
	}
	return []string{addr}
}
