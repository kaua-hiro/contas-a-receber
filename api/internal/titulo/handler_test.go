package titulo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func novoServidor(t *testing.T) http.Handler {
	t.Helper()
	h := NovoHandler(NovoMemoriaStore())
	h.agora = func() time.Time { return time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC) }
	mux := http.NewServeMux()
	h.Registrar(mux)
	return mux
}

func fazer(t *testing.T, srv http.Handler, metodo, caminho, corpo string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, caminho, strings.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestCriarTitulo(t *testing.T) {
	casos := []struct {
		nome   string
		corpo  string
		status int
	}{
		{"válido", `{"cliente":"Loja A","valor_centavos":15050,"vencimento":"2026-02-01"}`, http.StatusCreated},
		{"JSON malformado", `{"cliente":`, http.StatusBadRequest},
		{"campo desconhecido", `{"cliente":"A","valor_centavos":1,"vencimento":"2026-02-01","x":1}`, http.StatusBadRequest},
		{"cliente vazio", `{"cliente":"  ","valor_centavos":100,"vencimento":"2026-02-01"}`, http.StatusUnprocessableEntity},
		{"valor zero", `{"cliente":"A","valor_centavos":0,"vencimento":"2026-02-01"}`, http.StatusUnprocessableEntity},
		{"data inválida", `{"cliente":"A","valor_centavos":100,"vencimento":"01/02/2026"}`, http.StatusUnprocessableEntity},
		{"status inválido", `{"cliente":"A","valor_centavos":100,"vencimento":"2026-02-01","status":"X"}`, http.StatusUnprocessableEntity},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := fazer(t, novoServidor(t), http.MethodPost, "/titulos", c.corpo)
			if rec.Code != c.status {
				t.Fatalf("status = %d, esperado %d (corpo: %s)", rec.Code, c.status, rec.Body)
			}
		})
	}
}

func TestCriarDefineStatusAberto(t *testing.T) {
	rec := fazer(t, novoServidor(t), http.MethodPost, "/titulos",
		`{"cliente":"Loja A","valor_centavos":100,"vencimento":"2026-02-01"}`)

	var got Titulo
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 || got.Status != StatusAberto {
		t.Fatalf("título inesperado: %+v", got)
	}
}

func TestFluxoCompleto(t *testing.T) {
	srv := novoServidor(t)
	fazer(t, srv, http.MethodPost, "/titulos", `{"cliente":"Loja A","valor_centavos":1000,"vencimento":"2026-01-10"}`)
	fazer(t, srv, http.MethodPost, "/titulos", `{"cliente":"Loja B","valor_centavos":2000,"vencimento":"2026-03-10"}`)

	if rec := fazer(t, srv, http.MethodGet, "/titulos/1", ""); rec.Code != http.StatusOK {
		t.Fatalf("buscar: status %d", rec.Code)
	}
	if rec := fazer(t, srv, http.MethodGet, "/titulos/99", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("buscar inexistente: status %d", rec.Code)
	}
	if rec := fazer(t, srv, http.MethodGet, "/titulos/abc", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("id inválido: status %d", rec.Code)
	}

	rec := fazer(t, srv, http.MethodPut, "/titulos/2", `{"cliente":"Loja B2","valor_centavos":2500,"vencimento":"2026-03-10"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Loja B2") {
		t.Fatalf("atualizar: status %d corpo %s", rec.Code, rec.Body)
	}

	if rec := fazer(t, srv, http.MethodPatch, "/titulos/2/pagar", ""); rec.Code != http.StatusOK {
		t.Fatalf("pagar: status %d", rec.Code)
	}
	if rec := fazer(t, srv, http.MethodPatch, "/titulos/2/pagar", ""); rec.Code != http.StatusConflict {
		t.Fatalf("pagar duas vezes: status %d", rec.Code)
	}

	rec = fazer(t, srv, http.MethodGet, "/titulos/resumo", "")
	var r Resumo
	if err := json.NewDecoder(rec.Body).Decode(&r); err != nil {
		t.Fatal(err)
	}
	esperado := Resumo{Quantidade: 2, TotalAbertoCentavos: 1000, TotalPagoCentavos: 2500, QtdVencidos: 1, TotalVencidoCentavo: 1000}
	if r != esperado {
		t.Fatalf("resumo = %+v, esperado %+v", r, esperado)
	}

	if rec := fazer(t, srv, http.MethodDelete, "/titulos/1", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("remover: status %d", rec.Code)
	}
	if rec := fazer(t, srv, http.MethodDelete, "/titulos/1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("remover de novo: status %d", rec.Code)
	}
}

func TestListarComFiltros(t *testing.T) {
	srv := novoServidor(t)
	fazer(t, srv, http.MethodPost, "/titulos", `{"cliente":"Mercado Sol","valor_centavos":100,"vencimento":"2026-02-01"}`)
	fazer(t, srv, http.MethodPost, "/titulos", `{"cliente":"Padaria Lua","valor_centavos":200,"vencimento":"2026-02-01","status":"PAGO"}`)

	casos := []struct {
		query string
		qtd   int
	}{
		{"", 2},
		{"?status=PAGO", 1},
		{"?status=ABERTO", 1},
		{"?cliente=sol", 1},
		{"?cliente=xyz", 0},
	}
	for _, c := range casos {
		rec := fazer(t, srv, http.MethodGet, "/titulos"+c.query, "")
		var body struct {
			Titulos []Titulo `json:"titulos"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Titulos) != c.qtd {
			t.Errorf("%q: %d títulos, esperado %d", c.query, len(body.Titulos), c.qtd)
		}
	}

	if rec := fazer(t, srv, http.MethodGet, "/titulos?status=X", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("status inválido: %d", rec.Code)
	}
}
