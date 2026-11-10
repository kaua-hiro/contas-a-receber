package titulo

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Repositorio é o contrato de persistência. O handler depende dele, não da implementação,
// então trocar a memória por PostgreSQL não muda nada na camada HTTP.
type Repositorio interface {
	Listar(f Filtro) []Titulo
	Buscar(id int) (Titulo, error)
	Criar(e Entrada) Titulo
	Atualizar(id int, e Entrada) (Titulo, error)
	Remover(id int) error
	Pagar(id int) (Titulo, error)
	Resumo(hoje time.Time) Resumo
}

// MemoriaStore guarda os títulos em memória, protegido por mutex
// (cada requisição HTTP roda em uma goroutine diferente).
type MemoriaStore struct {
	mu      sync.RWMutex
	titulos map[int]Titulo
	proxID  int
	agora   func() time.Time
}

func NovoMemoriaStore() *MemoriaStore {
	return &MemoriaStore{titulos: map[int]Titulo{}, proxID: 1, agora: time.Now}
}

func (s *MemoriaStore) Listar(f Filtro) []Titulo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	busca := strings.ToLower(strings.TrimSpace(f.Cliente))
	resultado := []Titulo{}
	for _, t := range s.titulos {
		if f.Status != "" && t.Status != f.Status {
			continue
		}
		if busca != "" && !strings.Contains(strings.ToLower(t.Cliente), busca) {
			continue
		}
		resultado = append(resultado, t)
	}
	// map não tem ordem garantida: ordena para a resposta ser previsível.
	sort.Slice(resultado, func(i, j int) bool { return resultado[i].ID < resultado[j].ID })
	return resultado
}

func (s *MemoriaStore) Buscar(id int) (Titulo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.titulos[id]
	if !ok {
		return Titulo{}, ErrNaoEncontrado
	}
	return t, nil
}

func (s *MemoriaStore) Criar(e Entrada) Titulo {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Titulo{
		ID:            s.proxID,
		Cliente:       e.Cliente,
		ValorCentavos: e.ValorCentavos,
		Vencimento:    e.Vencimento,
		Status:        e.Status,
		CriadoEm:      s.agora().UTC(),
	}
	s.titulos[t.ID] = t
	s.proxID++
	return t
}

func (s *MemoriaStore) Atualizar(id int, e Entrada) (Titulo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.titulos[id]
	if !ok {
		return Titulo{}, ErrNaoEncontrado
	}
	t.Cliente = e.Cliente
	t.ValorCentavos = e.ValorCentavos
	t.Vencimento = e.Vencimento
	t.Status = e.Status
	s.titulos[id] = t
	return t, nil
}

func (s *MemoriaStore) Remover(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.titulos[id]; !ok {
		return ErrNaoEncontrado
	}
	delete(s.titulos, id)
	return nil
}

func (s *MemoriaStore) Pagar(id int) (Titulo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.titulos[id]
	if !ok {
		return Titulo{}, ErrNaoEncontrado
	}
	if t.Status == StatusPago {
		return Titulo{}, ErrJaPago
	}
	t.Status = StatusPago
	s.titulos[id] = t
	return t, nil
}

func (s *MemoriaStore) Resumo(hoje time.Time) Resumo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var r Resumo
	for _, t := range s.titulos {
		r.Quantidade++
		switch t.Status {
		case StatusAberto:
			r.TotalAbertoCentavos += t.ValorCentavos
		case StatusPago:
			r.TotalPagoCentavos += t.ValorCentavos
		}
		if t.Vencido(hoje) {
			r.QtdVencidos++
			r.TotalVencidoCentavo += t.ValorCentavos
		}
	}
	return r
}
