// Cliente HTTP da API. Centraliza URL base e tratamento de erro.
const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function request(path, options = {}) {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (res.status === 204) return null

  const body = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(body.erro ?? `Erro HTTP ${res.status}`)
  return body
}

export const api = {
  listar: (filtros = {}) => {
    const params = new URLSearchParams()
    if (filtros.status) params.set('status', filtros.status)
    if (filtros.cliente) params.set('cliente', filtros.cliente)
    const qs = params.toString()
    return request(`/titulos${qs ? `?${qs}` : ''}`).then((b) => b.titulos)
  },
  resumo: () => request('/titulos/resumo'),
  criar: (dados) => request('/titulos', { method: 'POST', body: JSON.stringify(dados) }),
  atualizar: (id, dados) => request(`/titulos/${id}`, { method: 'PUT', body: JSON.stringify(dados) }),
  pagar: (id) => request(`/titulos/${id}/pagar`, { method: 'PATCH' }),
  remover: (id) => request(`/titulos/${id}`, { method: 'DELETE' }),
}

export const formatarBRL = (centavos) =>
  (centavos / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export const formatarData = (iso) => iso.split('-').reverse().join('/')
