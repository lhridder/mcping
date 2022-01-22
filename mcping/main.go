package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-mc/mcping"
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
	Debug      bool     `json:"debug"`
	Delay      int      `json:"delay"`
}

var config Config
var protocol = flag.Int("p", 578, "The minecraft protocol version")

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

	jsonParser := json.NewDecoder(jsonFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		panic(err)
	}

	err = jsonFile.Close()
	if err != nil {
		panic(err)
	}

	fmt.Println("Started mcping with debug mode: " + strconv.FormatBool(config.Debug) + " with delay: " + strconv.Itoa(config.Delay))
	// Start prom listener
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(config.PromListen, nil)
		if err != nil {
			panic(err)
			return
		}
	}()
	// Start fetching targets
	for {
		if config.Debug {
			fmt.Println("Fetching all targets...")
		}
		for _, host := range config.Targets {
			if config.Debug {
				fmt.Println("Fetching " + host)
			}
			players, delay := getServerStats(host)
			playerCount.With(prometheus.Labels{"host": host}).Set(float64(players))
			pingDelay.With(prometheus.Labels{"host": host}).Set(float64(delay))
			if config.Debug {
				fmt.Println("Fetched " + host + ": " + strconv.Itoa(players))
			}
		}
		time.Sleep(time.Duration(config.Delay) * time.Second)
	}
}

// Get target stats
func getServerStats(host string) (playercount int, delay time.Duration) {
	addrs := lookupMC(host)
	if config.Debug {
		fmt.Println("Found addrs: " + strings.Join(addrs, ","))
	}
	for _, addr := range addrs {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Dial error for " + addr + ": ")
			fmt.Println(err)
			continue
		}
		hostname, _, _ := net.SplitHostPort(addr)
		status, delay, err := mcping.PingAndListConn(conn, *protocol, hostname)
		if err != nil {
			fmt.Println("Ping error for " + addr + ": ")
			fmt.Println(err)
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
