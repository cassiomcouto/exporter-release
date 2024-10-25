# Etapa 1: Build da aplicação
FROM golang:1.22-alpine AS builder

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de configuração e módulos do Go
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante do código do aplicativo
COPY . .

# Construir o binário do aplicativo
RUN CGO_ENABLED=0 GOOS=linux go build -o exporter-release ./cmd/exporter-release/main.go

# Etapa 2: Imagem final para execução
FROM alpine:latest

# Diretório de trabalho no container
WORKDIR /root/

# Copiar arquivos de configuração
COPY config/config.yaml config/repos_and_charts.yaml ./config/

# Copiar o binário do build para a imagem final
COPY --from=builder /app/exporter-release .

# Expor a porta configurada (altere se necessário)
EXPOSE 8080

# Executar o binário
CMD ["./exporter-release", "-config=config/config.yaml"]