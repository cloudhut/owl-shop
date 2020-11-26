# Owl Shop / Kafka Load Test Tool

Owl Shop is an imaginary ecommerce shop that simulates microservices exchanging data via Apache Kafka.
You can use this tool to load test Kafka clusters or fill it with demo data.

## What does it do?

Unlike common load testers Owl Shop tries to mimic the behaviour of an actual microservice landscape. That means it will
also send tombstones, messages with the same key on compacted topics and it also consumes messages from other topics
via a consumer group. 

**Produced topics:**

- ${globalPrefix}addresses
- ${globalPrefix}customers
- ${globalPrefix}frontend-events
- ${globalPrefix}orders

**Consumed topics:**

- ${globalPrefix}customers (AddressService, OrderService)

## Getting started

You can configure Owl Shop via arguments and and a YAML config. The main configuration should take place via the YAML
config. Via arguments you must specify the filepath to your YAML config and you can configure sensitive input via arguments
instead of putting them into the YAML file.

**Available flags:**

- `-config.filepath`
- `-shop.kafka.sasl.password`
- `-shop.kafka.sasl.gssapi.password`
- `-shop.kafka.tls.passphrase`

**YAML config:**

```yaml
shop:
  globalPrefix: owlshop- # Prefix to be used for clientID, consumergroupIDs and all topic names. Defaults to "owlshop-"
  traffic:
    pattern: constant # Defaults to constant. Currently this is the only supported pattern
    interval:
      rate: # The number of pageimpressions to simulate on the shop / the specified interval duration. This roughly equals to the number of Kafka messages beind produced
      duration: # Interval duration in which ${rate} page impressions shall be simulated (e.g. 500 impressions / 1s)
  kafka:
    brokers:
      - bootstrap-brokers.mycompany.com:9092
    sasl:
      enabled: true
      mechanism: PLAIN # PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI, OAUTHBEARER
      username: johndoe
      # password: set via flags
      # gssapi:
      #   authType:
      #   keyTabPath:
      #   kerberosConfigPath:
      #   serviceName:
      #   username:
      #   password: # can be set via the --kafka.sasl.gssapi.password flag as well
      #   realm:
    tls:
      enabled: true # Defaults to system's cert pool
      # caFilepath:
      # certFilepath:
      # keyFilepath:
      # passphrase: # This can be set via the --kafka.tls.passphrase flag as well
      # insecureSkipTlsVerify: false
    clientId: OwlShop

logger:
  level: info # Defaults to info. Valid values are: debug, info, warn, error, fatal
```
