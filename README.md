# CEP Race Condition

Sistema de consulta de endereço por CEP em Go que utiliza **multithreading** para consultar duas APIs simultaneamente e retornar apenas a resposta mais rápida.

---

## Índice

- [Sobre](#sobre)
- [Conceitos Aplicados](#conceitos-aplicados)
- [Arquitetura](#arquitetura)
- [Requisitos](#requisitos)
- [Instalação](#instalação)
- [Uso](#uso)
- [APIs Utilizadas](#apis-utilizadas)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Detalhes Técnicos](#detalhes-técnicos)
- [Licença](#licença)

---

## Sobre

Este projeto implementa um sistema de consulta de CEP que realiza requisições paralelas às APIs **BrasilAPI** e **ViaCEP**, aceitando apenas a resposta da API que responder mais rapidamente. O sistema possui um timeout global de **1 segundo** e trata cenários de falha de forma elegante.

### Funcionalidades Principais

- Requisições simultâneas (paralelas) às duas APIs
- Race condition — aceita apenas a resposta mais rápida
- Timeout de 1 segundo para ambas as APIs
- Tratamento de erros individual por API
- Saída formatada no terminal com identificação da API vencedora

---

## Conceitos Aplicados

| Conceito | Descrição |
|----------|-----------|
| **Goroutines** | Threads leves do Go para execução paralela |
| **Channels** | Comunicação segura entre goroutines |
| **Select** | Escolha não-determinística — aceita o primeiro resultado disponível |
| **Context** | Controle de timeout e cancelamento de requisições HTTP |
| **Buffered Channels** | Evita goroutine leaks ao descartar a resposta mais lenta |

---

## Arquitetura

```
┌─────────────────┐     ┌─────────────────┐
│   BrasilAPI     │     │    ViaCEP       │
│  (goroutine 1)  │     │  (goroutine 2)  │
│                 │     │                 │
│  HTTP Request   │     │  HTTP Request   │
│       ↓         │     │       ↓         │
│   Response      │     │   Response      │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────▼──────┐
              │  Channel    │
              │  (buffer 2) │
              └──────┬──────┘
                     │
              ┌──────▼──────┐
              │   select    │
              │  { case }   │
              └──────┬──────┘
                     │
         ┌───────────┼───────────┐
         │                       │
    ┌────▼────┐            ┌────▼────┐
    │ Vencedor│            │ Timeout │
    │  ← ch   │            │  ← 1s   │
    └─────────┘            └─────────┘
```

### Fluxo de Execução

1. O `main()` dispara duas goroutines independentes
2. Cada goroutine faz uma requisição HTTP para sua respectiva API
3. Ambas enviam o resultado para um **channel bufferizado** (`cap=2`)
4. O `select` bloqueia até que **uma** das operações esteja pronta
5. A primeira resposta a chegar é processada e exibida
6. A resposta mais lenta é descartada (o channel buffer evita leak)
7. Se nenhuma resposta chegar em 1 segundo, dispara **timeout**

---

## Requisitos

- [Go](https://golang.org/dl/) 1.21 ou superior
- Conexão com a internet

---

## Instalação

### Clone o repositório

```bash
git clone https://github.com/seu-usuario/cep-race-condition.git
cd cep-race-condition
```

### Verifique a instalação do Go

```bash
go version
# go version go1.21.x linux/amd64
```

---

## Uso

### Executar o projeto

```bash
go run main.go
```

### Exemplo de Saída (Sucesso)

```
Iniciando busca paralela para o CEP: 01310100

═══════════════════════════════════════════
   RESULTADO DA API MAIS RÁPIDA
═══════════════════════════════════════════
  API Vencedora: ViaCEP
───────────────────────────────────────────
  CEP:         01310-100
  Cidade:      São Paulo
  Estado:      SP
  Bairro:      Bela Vista
  Logradouro:  Avenida Paulista
  Complemento: 
═══════════════════════════════════════════
```

### Exemplo de Saída (Timeout)

```
Iniciando busca paralela para o CEP: 01310100

TIMEOUT: Nenhuma API respondeu dentro de 1 segundo.
```

### Personalizar o CEP

Edite a variável `cep` no `main()`:

```go
cep := "01001000"  // CEP da Praça da Sé, São Paulo
```

---

## APIs Utilizadas

| API | URL Base | Documentação |
|-----|----------|--------------|
| **BrasilAPI** | `https://brasilapi.com.br/api/cep/v1/{cep}` | [brasilapi.com.br](https://brasilapi.com.br) |
| **ViaCEP** | `http://viacep.com.br/ws/{cep}/json/` | [viacep.com.br](https://viacep.com.br) |

---

## Estrutura do Projeto

```
cep-race-condition/
├── main.go          # Código fonte principal
├── go.mod           # Módulo Go (gerado automaticamente)
└── README.md        # Este arquivo
```

---

## Detalhes Técnicos

### Goroutine Leak Prevention

O channel é criado com buffer de capacidade 2:

```go
ch := make(chan Result, 2)
```

Isso garante que a goroutine mais lenta possa enviar seu resultado sem bloquear indefinidamente, mesmo que o `select` já tenha escolhido o vencedor. Sem o buffer, a goroutine perdedora ficaria presa para sempre (goroutine leak).

### Timeout em Duas Camadas

| Camada | Mecanismo | Propósito |
|--------|-----------|-----------|
| **Individual** | `context.WithTimeout(1s)` | Cancela a requisição HTTP se a API demorar |
| **Global** | `time.After(1s)` no `select` | Limita o tempo total de espera pelo vencedor |

### Tratamento de Erros

Cada goroutine encapsula seus próprios erros (network, JSON decode, HTTP status) no struct `Result`, permitindo que o `select` trate apenas o vencedor. Se a API vencedora retornar erro, o sistema exibe a mensagem de erro em vez de crashar.
