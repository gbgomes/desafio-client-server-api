package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Nome string `json:"nome"`
	Bid  string `json:"bid"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao?moeda=USD", nil)
	if err != nil {
		str := "Erro ao fazer requisição: " + err.Error()
		log.Fatal(str)
	}

	res, err := http.DefaultClient.Do(req)
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		log.Fatal("Timeout no request ao server")
	}
	if err != nil {
		log.Fatal(err)
	}
	status := res.StatusCode
	if status == 500 {
		log.Fatal("Erro interno reportado pelo servidor de cotações")
	}
	if status == 408 {
		log.Fatal("Timeout reportado pelo servidor de cotações")
	}
	defer res.Body.Close()

	var moeda Cotacao
	res1, err := io.ReadAll(res.Body)
	err = json.Unmarshal(res1, &moeda)
	if err != nil {
		str := "Erro ao fazer parse da resposta: " + err.Error()
		log.Fatal(str)
	}
	fmt.Printf("Cotacao: %s, Cotação: %v", moeda.Nome, moeda.Bid)

	gravaResultado(moeda)
}

func gravaResultado(moeda Cotacao) {

	f, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write([]byte(moeda.Nome + ": " + moeda.Bid + "\n"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
}
