package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// https://mholt.github.io/json-to-go/ para converte json para stuc
type Moeda struct {
	Nome string  `json:"nome"`
	Bid  float64 `json:"bid,string"`
}

type Cotacao struct {
	Code       string    `json:"code,string"`
	Codein     string    `json:"codein,string"`
	Name       string    `json:"name,string"`
	High       float64   `json:"high,string"`
	Low        float64   `json:"low,string"`
	VarBid     float64   `json:"varBid,string"`
	PctChange  float64   `json:"pctChange,string"`
	Bid        float64   `json:"bid,string"`
	Ask        float64   `json:"ask,string"`
	Timestamp  string    `json:"timestamp,string"`
	CreateDate time.Time `json:"create_date,string"`
}

func main() {
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/cotacao", buscaCotacao)
	http.ListenAndServe(":8080", mux1)
}

func buscaCotacao(w http.ResponseWriter, r *http.Request) {
	// o método atualizaCotacoes poderia estar em um outro contexto/lugar
	// onde seria cahamado de tempos em tempos ou somente após o fechamanto do mercado
	// evitando a coleta de informações iguais a cada request do client
	moedas, err := atualizaCotacoes(r.Context())
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		log.Println("Timeout no processamanto da requisição")
		http.Error(w, "Timeout no processamanto da requisição", http.StatusRequestTimeout)
		return
	}
	if err != nil {
		http.Error(w, "Erro no processamanto da requisição", http.StatusInternalServerError)
		return
	}

	param := r.URL.Query().Get("moeda")
	var moeda Moeda
	for _, m := range moedas {
		if m.Code == param {
			moeda.Nome = m.Code
			moeda.Bid = m.Bid
			break
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(moeda)
}

func atualizaCotacoes(ctx context.Context) ([]Cotacao, error) {
	moedasDet, err := buscaCotacoes(ctx)
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		log.Println("Timeout no request da API de cotações")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	err = gravaCotacoes(ctx, moedasDet)
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		log.Println("Timeout no acesso ao Banco de Dados")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return moedasDet, err
}

func buscaCotacoes(ctx context.Context) ([]Cotacao, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	//ctx, cancel := context.WithTimeout(ctx, 2*time.Microsecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL,BRL-USD", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v", err)
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	res1, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var moedasDet []Cotacao
	var moedasReq map[string]any
	err = json.Unmarshal(res1, &moedasReq)
	if err != nil {
		return nil, err
	}

	for _, moeda := range moedasReq {
		a := moeda.(map[string]interface{})["ask"].(string)
		h := moeda.(map[string]interface{})["high"].(string)
		l := moeda.(map[string]interface{})["low"].(string)
		v := moeda.(map[string]interface{})["varBid"].(string)
		p := moeda.(map[string]interface{})["pctChange"].(string)
		b := moeda.(map[string]interface{})["bid"].(string)
		t := moeda.(map[string]interface{})["create_date"].(string)

		//pendente tratar cada erro de conversão
		af, _ := strconv.ParseFloat(a, 64)
		hf, _ := strconv.ParseFloat(h, 64)
		lf, _ := strconv.ParseFloat(l, 64)
		vf, _ := strconv.ParseFloat(v, 64)
		pf, _ := strconv.ParseFloat(p, 64)
		bf, _ := strconv.ParseFloat(b, 64)
		tt, _ := time.Parse("2006-01-02 15:04:05", t)

		m1 := Cotacao{
			Name:       moeda.(map[string]interface{})["name"].(string),
			Code:       moeda.(map[string]interface{})["code"].(string),
			Ask:        af,
			Codein:     moeda.(map[string]interface{})["codein"].(string),
			High:       hf,
			Low:        lf,
			VarBid:     vf,
			PctChange:  pf,
			Bid:        bf,
			Timestamp:  moeda.(map[string]interface{})["timestamp"].(string),
			CreateDate: tt,
		}
		moedasDet = append(moedasDet, m1)
	}
	return moedasDet, err
}

func gravaCotacoes(ctx context.Context, mds []Cotacao) error {

	//	CREATE TABLE cotacoes (
	//		code char(3) NOT NULL PRIMARY KEY,
	//		codein char(3),
	//		name varchar(40),
	//		high double,
	//		low double,
	//		varBid double,
	//		pctChange double,
	//		bid double,
	//		ask double,
	//		timestamp varchar(10),
	//		createDate datetime)

	db, err := sql.Open("mysql", "root:#Pk0cxh281513@tcp(localhost:3306)/goexpert?parseTime=true")
	if err != nil {
		return err
	}
	defer db.Close()

	for _, md := range mds {
		//coleta do BD a cotação da moeda nada data recebida da API
		md_bd, err := cotacaoByCode(db, ctx, md.Code)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		// se existir, verifica se é o mesmo valor recebido da API
		//Se for diferente atualiza o BD
		// esta versão mantêm apenas a ultima cotação, ou seja, não mantêm as cotações de várias datas
		if md_bd != nil && md_bd.Bid != md.Bid {
			err = atualizaCotacao(db, ctx, &md)
			if err != sql.ErrNoRows {
				return err
			}
		}

		//se não existir a cotação, insere no BD
		if md_bd == nil {
			err = insereCotacao(db, ctx, &md)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func cotacaoByCode(db *sql.DB, ctx context.Context, code string) (*Cotacao, error) {
	//ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("select code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate from cotacoes where code = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var c Cotacao
	err = stmt.QueryRowContext(ctx, code).Scan(&c.Code, &c.Codein, &c.Name, &c.High, &c.Low, &c.VarBid, &c.PctChange, &c.Bid, &c.Ask, &c.Timestamp, &c.CreateDate)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func insereCotacao(db *sql.DB, ctx context.Context, cotacao *Cotacao) error {
	//ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("insert into cotacoes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) " +
		"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, &cotacao.Code, &cotacao.Codein, &cotacao.Name, &cotacao.High, &cotacao.Low, &cotacao.VarBid, &cotacao.PctChange,
		&cotacao.Bid, &cotacao.Ask, &cotacao.Timestamp, &cotacao.CreateDate)
	if err != nil {
		return err
	}
	return err
}

func atualizaCotacao(db *sql.DB, ctx context.Context, cotacao *Cotacao) error {
	//ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("update cotacoes set codein = ?, name = ?, high = ?, low = ?, varBid = ?, pctChange = ?, bid = ?, " +
		" ask = ?, timestamp = ?, createDate = ? where code = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.Codein, cotacao.Name, cotacao.High, cotacao.Low, cotacao.VarBid, cotacao.PctChange,
		cotacao.Bid, cotacao.Ask, cotacao.Timestamp, cotacao.CreateDate, cotacao.Code)
	if err != nil {
		return err
	}
	return err
}
