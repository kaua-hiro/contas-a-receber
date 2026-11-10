<script setup>
import { reactive, watch } from 'vue'

// props: dados vindos do pai. emit: eventos enviados para o pai.
const props = defineProps({
  titulo: { type: Object, default: null },
})
const emit = defineEmits(['salvar', 'cancelar'])

const vazio = () => ({ cliente: '', valor: '', vencimento: '', status: 'ABERTO' })
const form = reactive(vazio())

// Quando o pai manda um título para editar, preenche o formulário.
watch(
  () => props.titulo,
  (t) => {
    Object.assign(
      form,
      t
        ? { cliente: t.cliente, valor: (t.valor_centavos / 100).toFixed(2), vencimento: t.vencimento, status: t.status }
        : vazio(),
    )
  },
  { immediate: true },
)

function enviar() {
  emit('salvar', {
    cliente: form.cliente.trim(),
    // Converte reais -> centavos com arredondamento para evitar 0.1 + 0.2.
    valor_centavos: Math.round(Number(form.valor) * 100),
    vencimento: form.vencimento,
    status: form.status,
  })
  if (!props.titulo) Object.assign(form, vazio())
}
</script>

<template>
  <form class="card form" @submit.prevent="enviar">
    <h2>{{ titulo ? `Editar título #${titulo.id}` : 'Novo título' }}</h2>
    <label>
      Cliente
      <input v-model="form.cliente" required maxlength="120" placeholder="Ex: Loja A" />
    </label>
    <label>
      Valor (R$)
      <input v-model="form.valor" type="number" min="0.01" step="0.01" required />
    </label>
    <label>
      Vencimento
      <input v-model="form.vencimento" type="date" required />
    </label>
    <label>
      Status
      <select v-model="form.status">
        <option value="ABERTO">Aberto</option>
        <option value="PAGO">Pago</option>
      </select>
    </label>
    <div class="acoes">
      <button type="submit">{{ titulo ? 'Salvar' : 'Cadastrar' }}</button>
      <button v-if="titulo" type="button" class="secundario" @click="emit('cancelar')">Cancelar</button>
    </div>
  </form>
</template>
