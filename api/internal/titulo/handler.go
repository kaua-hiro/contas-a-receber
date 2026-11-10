package titulo

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

const limiteCorpo = 1 << 20 // 1 MB

// Handler expõe o recurso /titulos via HTTP.
type Handler struct {
	repo  Repositorio
	agora func() time.Time
}

func NovoHandler(repo Repositorio) *Handler {
	return &Handler{repo: repo, agora: time.Now}
}

// Registrar adiciona as rotas no mux (padrões "MÉTODO /caminho/{param}" do Go 1.22+).
func (h *Handler) Registrar(mux *http.ServeMux) {
	mux.HandleFunc("GET /titulos", h.listar)
	mux.HandleFunc("GET /titulos/resumo", h.resumo)
	mux.HandleFunc("GET /titulos/{id}", h.buscar)
	mux.HandleFunc("POST /titulos", h.criar)
	mux.HandleFunc("PUT /titulos/{id}", h.atualizar)
	mux.HandleFunc("PATCH /titulos/{id}/pagar", h.pagar)
	mux.HandleFunc("DELETE /titulos/{id}", h.remover)
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := Filtro{Status: Status(q.Get("status")), Cliente: q.Get("cliente")}
	if f.Status != "" && f.Status != StatusAberto && f.Status != StatusPago {
		responderErro(w, http.StatusBadRequest, "status deve ser ABERTO ou PAGO")
		return
	}
	responderJSON(w, http.StatusOK, map[string]any{"titulos": h.repo.Listar(f)})
}

func (h *Handler) resumo(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, http.StatusOK, h.repo.Resumo(h.agora()))
}

func (h *Handler) buscar(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}
	t, err := h.repo.Buscar(id)
	if err != nil {
		responderErroDominio(w, err)
		return
	}
	responderJSON(w, http.StatusOK, t)
}

func (h *Handler) criar(w http.ResponseWriter, r *http.Request) {
	e, ok := lerEntrada(w, r)
	if !ok {
		return
	}
	responderJSON(w, http.StatusCreated, h.repo.Criar(e))
}

func (h *Handler) atualizar(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}
	e, ok := lerEntrada(w, r)
	if !ok {
		return
	}
	t, err := h.repo.Atualizar(id, e)
	if err != nil {
		responderErroDominio(w, err)
		return
	}
	responderJSON(w, http.StatusOK, t)
}

func (h *Handler) pagar(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}
	t, err := h.repo.Pagar(id)
	if err != nil {
		responderErroDominio(w, err)
		return
	}
	responderJSON(w, http.StatusOK, t)
}

func (h *Handler) remover(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Remover(id); err != nil {
		responderErroDominio(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func lerID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		responderErro(w, http.StatusBadRequest, "id inválido")
		return 0, false
	}
	return id, true
}

func lerEntrada(w http.ResponseWriter, r *http.Request) (Entrada, bool) {
	var e Entrada
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, limiteCorpo))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		responderErro(w, http.StatusBadRequest, "JSON inválido")
		return Entrada{}, false
	}
	if err := e.Validar(); err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return Entrada{}, false
	}
	return e, true
}

func responderErroDominio(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		responderErro(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrJaPago):
		responderErro(w, http.StatusConflict, err.Error())
	default:
		log.Printf("erro inesperado: %v", err)
		responderErro(w, http.StatusInternalServerError, "erro interno")
	}
}

func responderErro(w http.ResponseWriter, status int, msg string) {
	responderJSON(w, status, map[string]string{"erro": msg})
}

func responderJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		log.Printf("erro ao escrever resposta: %v", err)
	}
}
