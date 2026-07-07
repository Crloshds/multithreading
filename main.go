package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Estruturas para mapear as respostas das APIs
type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

// Result encapsula o resultado de uma consulta
type Result struct {
	API     string
	Data    interface{}
	Error   error
}

// consultaBrasilAPI faz a requisição para a BrasilAPI
func consultaBrasilAPI(cep string, ch chan<- Result) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Error: err}
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Error: err}
		return
	}
	defer resp.Body.Close()

	var data BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- Result{API: "BrasilAPI", Error: err}
		return
	}

	ch <- Result{API: "BrasilAPI", Data: data}
}

// consultaViaCEP faz a requisição para a ViaCEP
func consultaViaCEP(cep string, ch chan<- Result) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		ch <- Result{API: "ViaCEP", Error: err}
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ch <- Result{API: "ViaCEP", Error: err}
		return
	}
	defer resp.Body.Close()

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- Result{API: "ViaCEP", Error: err}
		return
	}

	ch <- Result{API: "ViaCEP", Data: data}
}

// buscaEnderecoRapido inicia as duas requisições em paralelo e retorna a mais rápida
func buscaEnderecoRapido(cep string) {
	ch := make(chan Result, 2) // Buffer para evitar goroutine leak

	// Dispara as duas requisições simultaneamente (goroutines)
	go consultaBrasilAPI(cep, ch)
	go consultaViaCEP(cep, ch)

	// Timeout global de 1 segundo para a corrida
	timeout := time.After(1 * time.Second)

	select {
	case result := <-ch:
		if result.Error != nil {
			fmt.Printf("❌ Erro na API %s: %v\n", result.API, result.Error)
			return
		}

		fmt.Println("═══════════════════════════════════════════")
		fmt.Println("  RESULTADO DA API MAIS RÁPIDA")
		fmt.Println("═══════════════════════════════════════════")
		fmt.Printf("  API Vencedora: %s\n", result.API)
		fmt.Println("───────────────────────────────────────────")

		switch data := result.Data.(type) {
		case BrasilAPIResponse:
			fmt.Printf("  CEP:         %s\n", data.Cep)
			fmt.Printf("  Cidade:      %s\n", data.City)
			fmt.Printf("  Estado:      %s\n", data.State)
			fmt.Printf("  Bairro:      %s\n", data.Neighborhood)
			fmt.Printf("  Logradouro:  %s\n", data.Street)
			fmt.Printf("  Serviço:     %s\n", data.Service)
		case ViaCEPResponse:
			fmt.Printf("  CEP:         %s\n", data.Cep)
			fmt.Printf("  Cidade:      %s\n", data.Localidade)
			fmt.Printf("  Estado:      %s\n", data.Uf)
			fmt.Printf("  Bairro:      %s\n", data.Bairro)
			fmt.Printf("  Logradouro:  %s\n", data.Logradouro)
			fmt.Printf("  Complemento: %s\n", data.Complemento)
		}
		fmt.Println("═══════════════════════════════════════════")

	case <-timeout:
		fmt.Println("TIMEOUT: Nenhuma API respondeu dentro de 1 segundo.")
	}
}

func main() {
	// CEP de exemplo: Mcdonalds em Brasilia
	cep := "71560-100"

	fmt.Printf("\n Iniciando busca paralela para o CEP: %s\n\n", cep)
	buscaEnderecoRapido(cep)
}
