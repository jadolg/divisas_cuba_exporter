package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

type APIResponse struct {
	Currency          string    `json:"currency"`
	ExchangeDirection string    `json:"exchange_direction"`
	DateTime          time.Time `json:"date_time"`
	Rates             struct {
		Usd struct {
			Buy  float64 `json:"buy"`
			Sell float64 `json:"sell"`
			Mid  float64 `json:"mid"`
		} `json:"USD"`
		Mlc struct {
			Buy  float64 `json:"buy"`
			Sell float64 `json:"sell"`
			Mid  float64 `json:"mid"`
		} `json:"MLC"`
		Cup struct {
			Buy  int `json:"buy"`
			Sell int `json:"sell"`
			Mid  int `json:"mid"`
		} `json:"CUP"`
		Eur struct {
			Buy  float64 `json:"buy"`
			Sell float64 `json:"sell"`
			Mid  float64 `json:"mid"`
		} `json:"EUR"`
	} `json:"rates"`
}

var (
	usdBuy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "usd_buy",
		Help: "Buy value in CUP of 1 USD",
	})
	usdSell = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "usd_sell",
		Help: "Sell value in CUP of 1 USD",
	})
	eurBuy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eur_buy",
		Help: "Buy value in CUP of 1 EUR",
	})
	eurSell = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eur_sell",
		Help: "Sell value in CUP of 1 EUR",
	})
	mlcBuy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mlc_buy",
		Help: "Buy value in CUP of 1 MLC",
	})
	mlcSell = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mlc_sell",
		Help: "Sell value in CUP of 1 MLC",
	})
)

func getCurrentExchangeRates() (*APIResponse, error) {
	url := "https://exchange-rate-api.pages.dev/api/v2/informal/target/cup.json"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Divisas Exporter")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	data := APIResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func getRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	return mux

}

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6869"
	}
	router := getRouter()
	log.Infof("Starting server at port %s\n", port)
	log.Info("Metrics available at /metrics")
	log.Info("Health check available at /health")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatal(err)
	}
}

func main() {
	go startServer()

	for {
		log.Info("Updating exchange rates")
		rates, err := getCurrentExchangeRates()
		if err != nil {
			log.Fatal(err)
		}

		usdBuy.Set(rates.Rates.Usd.Buy)
		usdSell.Set(rates.Rates.Usd.Sell)
		eurBuy.Set(rates.Rates.Eur.Buy)
		eurSell.Set(rates.Rates.Eur.Sell)
		mlcBuy.Set(rates.Rates.Mlc.Buy)
		mlcSell.Set(rates.Rates.Mlc.Sell)

		log.Info("Exchange rates updated. Waiting for 2 hours")
		time.Sleep(2 * time.Hour)
	}
}
