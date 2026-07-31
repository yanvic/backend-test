# Por que fiz assim

## Go, não Node

Simplicidade da linguagem e pelo modelo de concorrência — goroutines
leves, channels, servidor HTTP que já dispara uma goroutine por request.
Sem framework pesado, sem ceremony.

## Clean Architecture + DDD leve

Domínio (`internal/domain`) não importa framework nenhum — só `time`.
As regras de negócio são funções puras, testáveis sem subir servidor
nem banco. O use case orquestra, o handler só traduz HTTP ↔ domínio.

[Exemplo prático de Clean Architecture em Go](https://threedots.tech/post/introducing-clean-architecture/)

## SQLite com interface PartRepository

Trocar de banco é só implementar a interface `PartRepository`. O domínio
nem sabe que banco existe — funciona com SQLite, memória, ou qualquer outro.

Tem implementação em memória pros testes unitários (rápido, sem I/O) e
SQLite pra produção.

## Value Objects tipados

`Stock`, `CriticalityLevel`, `LeadTimeDays`, `DailySales` são tipos nomeados,
não `int`/`float64` crus. O compilador pega se eu passar `leadTimeDays` onde
se espera `minimumStock`. Type safety sem overhead — custo zero em runtime.

[Value Objects como named types em Go](https://programmingpercy.tech/blog/ddd-value-objects-in-go/)

## GET /restock/priorities retorna tudo, não filtra

O endpoint pede "prioridades", não "quem precisa repor". É uma visão geral
ordenada por urgência, com tie-break: criticalityLevel → averageDailySales → nome.

## Validação no use case, não no handler

Se a validação estivesse no handler HTTP, qualquer outro meio de entrada
(CLI, gRPC, mensageria) teria que reimplementar. Use case é o dono das
invariantes do negócio.

## PubSub não, mas a porta tá aberta

O desafio proíbe API externa, então nada de Google PubSub. Mas deixei a
fronteira pronta: uma interface `EventPublisher` poderia ser plugada
depois sem mexer em regra de negócio.

## UUID automático

Cliente não manda `id` → servidor gera UUID v4. Manda → usa o que veio.
Flexível pra integração com sistemas legados.

## Mensagens em português

Todas as mensagens de erro e validação tão centralizadas em
`internal/domain/messages.go`. Mudar um texto = editar um arquivo só.

## Fuzz testing + benchmarks

30 segundos de fuzz no cálculo de urgência, com invariantes verificadas
a cada iteração (nunca panica, criticality=0 ⇒ urgency=0, determinístico).

Benchmarks com 100, 1.000 e 10.000 peças provam que o `sort.Slice`
O(n log n) escala dentro do esperado.

[Fuzzing nativo do Go](https://go.dev/doc/security/fuzz/)

## Concorrência

O repositório usa `sync.RWMutex` — múltiplas leituras simultâneas sem
contenção, escrita exclusiva. Teste de concorrência dispara 50 goroutines
de escrita + 25 de leitura ao mesmo tempo, `-race` limpo.
