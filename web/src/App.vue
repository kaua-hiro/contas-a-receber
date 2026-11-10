<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { api, formatarBRL, formatarData } from './api'
import TituloForm from './components/TituloForm.vue'

const titulos = ref([])
const resumo = ref(null)
const filtroStatus = ref('')
const filtroCliente = ref('')
const editando = ref(null)
const carregando = ref(false)
const erro = ref('')

const hoje = new Date().toISOString().slice(0, 10)
const estaVencido = (t) => t.status === 'ABERTO' && t.vencimento < hoje

// computed: recalcula sozinho quando "titulos" muda.
const totalListado = computed(() => titulos.value.reduce((soma, t) => soma + t.valor_centavos, 0))

async function carregar() {
  carregando.value = true
  erro.value = ''
  try {
    const [lista, r] = await Promise.all([
      api.listar({ status: filtroStatus.value, cliente: filtroCliente.value }),
      api.resumo(),
    ])
    titulos.value = lista
    resumo.value = r
  } catch (e) {
    erro.value = e.message
  } finally {
    carregando.value = false
  }
}

// Executa uma ação e recarrega a lista; mostra o erro da API se falhar.
async function executar(acao) {
  erro.value = ''
  try {
    await acao()
    await carregar()
  } catch (e) {
    erro.value = e.message
  }
}

function salvar(dados) {
  executar(async () => {
    if (editando.value) await api.atualizar(editando.value.id, dados)
    else await api.criar(dados)
    editando.value = null
  })
}

function remover(t) {
  if (confirm(`Remover o título de ${t.cliente}?`)) executar(() => api.remover(t.id))
}

// Debounce simples na busca por cliente para não disparar uma requisição por tecla.
let timer
watch(filtroCliente, () => {
  clearTimeout(timer)
  timer = setTimeout(carregar, 300)
})
watch(filtroStatus, carregar)

onMounted(carregar)
</script>

<template>
  <main class="container">
    <header>
      <h1>Contas a Receber</h1>
      <p>CRUD de títulos com Go + Vue</p>
    </header>

    <section v-if="resumo" class="resumo">
      <div class="card"><span>Em aberto</span><strong>{{ formatarBRL(resumo.total_aberto_centavos) }}</strong></div>
      <div class="card"><span>Recebido</span><strong>{{ formatarBRL(resumo.total_pago_centavos) }}</strong></div>
      <div class="card alerta">
        <span>Vencidos ({{ resumo.qtd_vencidos }})</span><strong>{{ formatarBRL(resumo.total_vencido_centavos) }}</strong>
      </div>
    </section>

    <p v-if="erro" class="erro" role="alert">{{ erro }}</p>

    <div class="grid">
      <TituloForm :titulo="editando" @salvar="salvar" @cancelar="editando = null" />

      <section class="card">
        <div class="filtros">
          <input v-model="filtroCliente" placeholder="Buscar cliente..." />
          <select v-model="filtroStatus">
            <option value="">Todos</option>
            <option value="ABERTO">Abertos</option>
            <option value="PAGO">Pagos</option>
          </select>
        </div>

        <p v-if="carregando">Carregando...</p>
        <p v-else-if="titulos.length === 0" class="vazio">Nenhum título encontrado.</p>

        <table v-else>
          <thead>
            <tr><th>#</th><th>Cliente</th><th>Valor</th><th>Vencimento</th><th>Status</th><th></th></tr>
          </thead>
          <tbody>
            <tr v-for="t in titulos" :key="t.id" :class="{ vencido: estaVencido(t) }">
              <td>{{ t.id }}</td>
              <td>{{ t.cliente }}</td>
              <td>{{ formatarBRL(t.valor_centavos) }}</td>
              <td>{{ formatarData(t.vencimento) }}</td>
              <td>
                <span :class="['badge', t.status.toLowerCase()]">{{ estaVencido(t) ? 'VENCIDO' : t.status }}</span>
              </td>
              <td class="acoes">
                <button v-if="t.status === 'ABERTO'" @click="executar(() => api.pagar(t.id))">Pagar</button>
                <button class="secundario" @click="editando = t">Editar</button>
                <button class="perigo" @click="remover(t)">Excluir</button>
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr><td colspan="2">Total listado</td><td colspan="4">{{ formatarBRL(totalListado) }}</td></tr>
          </tfoot>
        </table>
      </section>
    </div>
  </main>
</template>
