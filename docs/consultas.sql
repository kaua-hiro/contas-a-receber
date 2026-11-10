-- Modelo relacional equivalente ao domínio da API (PostgreSQL).
CREATE TABLE clientes (
  id    SERIAL PRIMARY KEY,
  nome  VARCHAR(120) NOT NULL
);

CREATE TABLE titulos (
  id              SERIAL PRIMARY KEY,
  cliente_id      INT NOT NULL REFERENCES clientes(id),
  valor_centavos  BIGINT NOT NULL CHECK (valor_centavos > 0),
  vencimento      DATE NOT NULL,
  status          VARCHAR(10) NOT NULL CHECK (status IN ('ABERTO', 'PAGO')),
  criado_em       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 1. Títulos em aberto com o nome do cliente (JOIN + WHERE)
SELECT t.id, c.nome, t.valor_centavos, t.vencimento
FROM titulos t
JOIN clientes c ON c.id = t.cliente_id
WHERE t.status = 'ABERTO'
ORDER BY t.vencimento;

-- 2. Totais por status (equivalente ao GET /titulos/resumo)
SELECT status, COUNT(*) AS quantidade, SUM(valor_centavos) AS total_centavos
FROM titulos
GROUP BY status;

-- 3. Clientes com mais de R$ 1.000,00 em aberto (GROUP BY + HAVING)
SELECT c.nome, SUM(t.valor_centavos) AS total_aberto_centavos
FROM titulos t
JOIN clientes c ON c.id = t.cliente_id
WHERE t.status = 'ABERTO'
GROUP BY c.nome
HAVING SUM(t.valor_centavos) > 100000
ORDER BY total_aberto_centavos DESC;

-- 4. Títulos vencidos
SELECT id, cliente_id, valor_centavos, vencimento
FROM titulos
WHERE status = 'ABERTO' AND vencimento < CURRENT_DATE;

-- 5. Clientes sem nenhum título (LEFT JOIN)
SELECT c.id, c.nome
FROM clientes c
LEFT JOIN titulos t ON t.cliente_id = c.id
WHERE t.id IS NULL;
