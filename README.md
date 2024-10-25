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
  port: 8080
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

## Métricas Expostas

O exporter expõe as seguintes métricas no endpoint configurado (por padrão, `localhost:8080/metrics`):

- **`chart_release_version`**: Mostra a versão mais recente de um chart Helm e a data de disponibilização. Exemplo de métrica:

  ```plaintext
  chart_release_version{repo="https://grafana.github.io/helm-charts", chart="grafana", version="6.1.3", release_date="19-10-2024"} 1
  chart_release_version{repo="https://grafana.github.io/helm-charts", chart="loki", version="2.3.0", release_date="15-09-2024"} 1
  ```

## Usando o Exporter-Release com Docker

### Construindo a Imagem Docker

Para construir a imagem Docker do `exporter-release`, execute:

```bash
docker build -t exporter-release .
```

### Executando o Container com Configurações Externas

Para executar o `exporter-release` com arquivos de configuração externos, utilize o comando abaixo. Esse comando monta o diretório `config` do host no container, permitindo que os arquivos de configuração sejam lidos diretamente sem precisar copiá-los para a imagem.

```bash
docker run -p 8080:8080 -v $(pwd)/config:/root/config -e CONFIG_PATH=/root/config exporter-release
```

#### Explicação dos Parâmetros

- **`-p 8080:8080`**: Mapeia a porta 8080 do container para a porta 8080 do host. Substitua por outra porta, se necessário.
- **`-v $(pwd)/config:/root/config`**: Monta o diretório `config` do host no container, onde estão os arquivos `config.yaml` e `repos_and_charts.yaml`.
- **`-e CONFIG_PATH=/root/config`**: Define a variável de ambiente `CONFIG_PATH` para apontar para o diretório de configuração montado no container.
