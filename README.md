# Exporter Release

`exporter-release` é um exporter personalizado para Prometheus, criado para monitorar as versões mais recentes de charts em repositórios Helm. Ele expõe informações como a versão e a data de disponibilização de cada release, permitindo um acompanhamento contínuo das atualizações.

## Funcionalidades

- Monitora repositórios Helm configurados para obter informações de versão e data de release dos charts.
- Exposição de métricas no formato Prometheus com as seguintes labels:
  - `repo`: URL do repositório Helm.
  - `chart`: Nome do chart Helm.
  - `version`: Versão mais recente da release.
  - `release_date`: Data da disponibilização da release no formato `DD-MM-YYYY`.

## Instalação

1. Clone este repositório:

   ```bash
   git clone <url-do-repositorio>
   cd exporter-release
   ```

2.	Instale as dependências do projeto (caso não tenha o Go configurado):
   ```bash
   go mod tidy
   ```

3.	Compile e execute o projeto:
   ```bash
   go run main.go
   ```


## Configuração

### Arquivo `config.yaml`

O `config.yaml` é utilizado para definir a porta do servidor, o caminho para as métricas e o intervalo de verificação.

Exemplo de `config.yaml`:

```yaml
server:
  port: 8000
  metrics_path: "/metrics"
  check_interval: "5m"  # Intervalo de verificação em minutos
```

### Arquivo repos_and_charts.yaml

O `repos_and_charts.yaml` lista os repositórios Helm e os charts específicos a serem monitorados.

Exemplo de `repos_and_charts.yaml`:

```yaml
repositories:
  - url: "https://grafana.github.io/helm-charts"
    charts:
      - grafana
      - loki
  - url: "https://prometheus-community.github.io/helm-charts"
    charts:
      - prometheus
```
