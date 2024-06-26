# Desafio OpenTelemetry - Sistema de Temperatura por CEP

Este repositório foi criado exclusivamente para hospedar o código do desenvolvimento do Desfio de implementação do OpenTelemetry no Sistema de temperatura por CEP da **Pós Go Expert**, ministrado pela **Full Cycle**.

## Descrição do Desafio

A seguir estão os dados fornecidos na descrição do desafio.


### Objetivo

Desenvolver um sistema em Go que receba um CEP, identifica a cidade e retorna o clima atual (temperatura em graus celsius, fahrenheit e kelvin) juntamente com a cidade. Esse sistema deverá implementar OTEL(Open Telemetry) e Zipkin.

Basedo no cenário conhecido "Sistema de temperatura por CEP" denominado **Serviço B**, será incluso um novo projeto, denominado **Serviço A**.

### Requisitos - Serviço A (responsável pelo input)

- O sistema deve receber um input de 8 dígitos via POST, através do schema:  `{ "cep": "29902555" }`

- O sistema deve validar se o input é valido (contem 8 dígitos) e é uma STRING
    - Caso seja válido, será encaminhado para o **Serviço B** via HTTP
    
    - Caso não seja válido, deve retornar:
        - Código HTTP: **422**

        - Mensagem: **invalid zipcode**

### Requisitos - Serviço B (responsável pela orquestração)

- O sistema deve receber um CEP válido de 8 digitos

- O sistema deve realizar a pesquisa do CEP e encontrar o nome da localização, a partir disso, deverá retornar as temperaturas e formata-lás em: Celsius, Fahrenheit, Kelvin juntamente com o nome da localização.

- O sistema deve responder adequadamente nos seguintes cenários:
    - Em caso de sucesso:
        - Código HTTP: **200**

        - Response Body: `{ "city: "São Paulo", "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }`
    - Em caso de falha, caso o CEP não seja válido (com formato correto):

        - Código HTTP: **422**

        - Mensagem: **invalid zipcode**
    - Em caso de falha, caso o CEP não seja encontrado:
        - Código HTTP: **404**

        - Mensagem: **can not find zipcode**


### OTEL + Zipkin

Após a implementação dos serviços, adicione a implementação do OTEL + Zipkin:

- Implementar tracing distribuído entre Serviço A - Serviço B

- Utilizar span para medir o tempo de resposta do serviço de busca de CEP e busca de temperatura

### Dicas

- Utilize a API viaCEP (ou similar) para encontrar a localização que deseja consultar a temperatura: https://viacep.com.br/

- Utilize a API WeatherAPI (ou similar) para consultar as temperaturas desejadas: https://www.weatherapi.com/

- Para realizar a conversão de Celsius para Fahrenheit, utilize a seguinte fórmula: F = C * 1,8 + 32

- Para realizar a conversão de Celsius para Kelvin, utilize a seguinte fórmula: K = C + 273
    - Sendo F = Fahrenheit
    - Sendo C = Celsius
    - Sendo K = Kelvin
- Para dúvidas da implementação do OTEL, você pode [clicar aqui](https://opentelemetry.io/docs/languages/go/getting-started/)
- Para implementação de spans, você pode [clicar aqui](https://opentelemetry.io/docs/languages/go/instrumentation/#creating-spans)
- Você precisará utilizar um serviço de [collector do OTEL](https://opentelemetry.io/docs/collector/quick-start/)
- Para mais informações sobre Zipkin, você pode [clicar aqui](https://zipkin.io/)


### Entrega


- O código-fonte completo da implementação.

- Documentação explicando como rodar o projeto em ambiente dev.

- Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.


## Execução do Desafio


Foram criados dois módulos (um para cada serviço) chamados **service-a** e **sevice-b**. Dessa forma, é necesário excutar o comando `go mod tidy` em ambos os diretŕoios para que as dependências sejam baixadas. Abaixo está um comando para ser executado na raiz do projeto que realizará o download dos módulos em ambos os projetos:

```bash

wander@bsnote283:~/desafio-opentelemetry$ for caminhoModulo in $(find . -name go.mod); do cd $(dirname ${caminhoModulo}); go mod tidy; cd - ; done
/home/wander/desafio-opentelemetry
/home/wander/desafio-opentelemetry
wander@bsnote283:~/desafio-opentelemetry$ 

```

Após o download das dependências, basta utilizar o comando `docker-compose up --build -d` na raiz do projeto que serão geradas as imagens e, em seguida, os containers serão iniciados. Abaixo segue um exemplo dos containers em execução:

```bash

wander@bsnote283:~/desafio-opentelemetry$ docker ps
CONTAINER ID   IMAGE                                 COMMAND                  CREATED              STATUS                        PORTS                                                                                                                       NAMES
14f84e9abdb4   desafio-opentelemetry-service-a       "./server"               About a minute ago   Up About a minute             0.0.0.0:8181->8181/tcp, :::8181->8181/tcp                                                                                   service-a
36d565334530   desafio-opentelemetry-service-b       "./server"               About a minute ago   Up About a minute             0.0.0.0:8282->8282/tcp, :::8282->8282/tcp                                                                                   service-b
aa1309449490   openzipkin/zipkin                     "start-zipkin"           About a minute ago   Up About a minute (healthy)   9410/tcp, 0.0.0.0:9411->9411/tcp, :::9411->9411/tcp                                                                         zipkin
f31368b32a0b   otel/opentelemetry-collector:latest   "/otelcol --config=/…"   About a minute ago   Up About a minute             0.0.0.0:4317->4317/tcp, :::4317->4317/tcp, 0.0.0.0:8888-8889->8888-8889/tcp, :::8888-8889->8888-8889/tcp, 55678-55679/tcp   desafio-opentelemetry-otel-collector-1
wander@bsnote283:~/desafio-opentelemetry$ 

```
Serão inicializados quatro containers: service-a, service-b, zipkin e opentelemetry-collector. Após este ponto, já podem ser iniciados os testes. As portas de acesso dos containers estão descritas a seguir:

- service-a: porta **8181**

- service-a: porta **8282**

- zipkin: porta **9411**


### Testes do Webserver


Para a realização dos testes, basta executar o seguinte comando `go test ./...` a partir da raiz dos módulos. Abaixo segue o exemplo da execução:


```bash

wander@bsnote283:~/desafio-opentelemetry/service-a$ go test ./... 
?   	github.com/wandermaia/desafio-temperatura-cep/service-a/cmd/server	[no test files]
ok  	github.com/wandermaia/desafio-temperatura-cep/service-a/internal/infra/webserver/handlers	0.003s
wander@bsnote283:~/desafio-opentelemetry/service-a$ 
wander@bsnote283:~/desafio-opentelemetry/service-a$ cd ../service-b
wander@bsnote283:~/desafio-opentelemetry/service-b$ 
wander@bsnote283:~/desafio-opentelemetry/service-b$ go test ./... 
?   	github.com/wandermaia/desafio-temperatura-cep/service-b/cmd/server	[no test files]
ok  	github.com/wandermaia/desafio-temperatura-cep/service-b/internal/infra/webserver/handlers	0.841s
wander@bsnote283:~/desafio-opentelemetry/service-b$ 

```

Também foram criados os arquivos `service-a/api/apis_temperatura_service_a.http` e `service-a/api/apis_temperatura_service_b.http` para que os endpoints possam ser testados diretamente a partir do VScode. 

Para a geração do trace distribuído, deve ser utilizado apenas o arquivo `service-a/api/apis_temperatura_service_a.http`, pois ele realiza a chamada do service-a, que internamente chama o service-b.

Neste ponto é importante realizar algumas chamadas de testes, pois elas serão visualizadas no zipkin. A seguir estão alguns exemplos de chadas realizadas diretamente no VScode utilizando o arquivo `service-a/api/apis_temperatura_service_a.http`:

![chamada_http.png](/.img/chamada_http.png)


![chamada_http_invalida.png](/.img/chamada_http_invalida.png)


### Visualização dos Traces no Zipkin


Após a realização de algumas chamadas, basta acessar o endereço http://localhost:9411 que os traces poderão visualizados na interface web do Zipkin. Abaixo seguem os prints que demonstram as visualizações dos traces na interface web do Zipkin:


![zipkin01.png](/.img/zipkin01.png)


![zipkin02.png](/.img/zipkin02.png)


![zipkin03.png](/.img/zipkin03.png)

