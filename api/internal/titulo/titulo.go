// Package titulo contém o domínio de títulos a receber (cobranças).
package titulo

import (
	"errors"
	"strings"
	"time"
)

// Status representa a situação de um título.
type Status string

const (
	StatusAberto Status = "ABERTO"
	StatusPago   Status = "PAGO"
)

// LayoutData é o formato de data aceito na API (ISO 8601, só a data).
const LayoutData = "2006-01-02"

var (
	ErrNaoEncontrado = errors.New("título não encontrado")
	ErrJaPago        = errors.New("título já está pago")
)

// Titulo é uma cobrança a receber de um cliente.
// O valor é guardado em centavos (int64) para evitar erro de arredondamento de float.
type Titulo struct {
	ID            int       `json:"id"`
	Cliente       string    `json:"cliente"`
	ValorCentavos int64     `json:"valor_centavos"`
	Vencimento    string    `json:"vencimento"`
	Status        Status    `json:"status"`
	CriadoEm      time.Time `json:"criado_em"`
}

// Entrada é o corpo aceito em POST e PUT.
type Entrada struct {
	Cliente       string `json:"cliente"`
	ValorCentavos int64  `json:"valor_centavos"`
	Vencimento    string `json:"vencimento"`
	Status        Status `json:"status"`
}

// Validar normaliza a entrada e devolve erro se algum campo for inválido.
func (e *Entrada) Validar() error {
	e.Cliente = strings.TrimSpace(e.Cliente)
	if e.Cliente == "" {
		return errors.New("cliente é obrigatório")
	}
	if len(e.Cliente) > 120 {
		return errors.New("cliente deve ter no máximo 120 caracteres")
	}
	if e.ValorCentavos <= 0 {
		return errors.New("valor_centavos deve ser maior que zero")
	}
	if _, err := time.Parse(LayoutData, e.Vencimento); err != nil {
		return errors.New("vencimento deve estar no formato AAAA-MM-DD")
	}
	if e.Status == "" {
		e.Status = StatusAberto
	}
	if e.Status != StatusAberto && e.Status != StatusPago {
		return errors.New("status deve ser ABERTO ou PAGO")
	}
	return nil
}

// Vencido indica se o título está em aberto com vencimento anterior a hoje.
func (t Titulo) Vencido(hoje time.Time) bool {
	return t.Status == StatusAberto && t.Vencimento < hoje.Format(LayoutData)
}

// Filtro define os critérios opcionais da listagem.
type Filtro struct {
	Status  Status
	Cliente string
}

// Resumo agrega os títulos (equivalente a um GROUP BY status).
type Resumo struct {
	Quantidade          int   `json:"quantidade"`
	TotalAbertoCentavos int64 `json:"total_aberto_centavos"`
	TotalPagoCentavos   int64 `json:"total_pago_centavos"`
	QtdVencidos         int   `json:"qtd_vencidos"`
	TotalVencidoCentavo int64 `json:"total_vencido_centavos"`
}
