# publisher-subcriber

The following project is an illustration of using publisher subscriber model to create multiple helm releases. The request to install helm charts are sent to a topic `helm_installations` and request to uninstall helm charts are sent to topic `helm_deletions`. The consumer subscribes to these topics, and  creates and uninstalls helm releases for the message sent to it.

We are going to need a Kafka Cluster for our client application to operate with. We do have a file called as `docker-compose.yml`[./docker-compose.yml]. This is used to setup kafka cluster locally as a docker deamon.

In order to create the kafka cluster we can do this by using command:

```
docker compose up -d
```

# Creating a topic

We need to publish the request to install the helm charts. The request consists of `chart_url`,`namespace` and `values` in JSON format.This helm request has to be sent to a topic. In order to create a topic you can make use of following command:


We need to create two topics `helm_installations` and `helm_deletions`

```
docker compose exec broker \
  kafka-topics --create \
    --topic helm_installations \
    --bootstrap-server localhost:9092 \
    --replication-factor 1 \
    --partitions 1
```

```
docker compose exec broker \
  kafka-topics --create \
    --topic helm_deletions \
    --bootstrap-server localhost:9092 \
    --replication-factor 1 \
    --partitions 1
```

You can list the topic too once created using:

```
docker compose exec broker \
  kafka-topics --list\
    --bootstrap-server localhost:9092
```


Once the topic get's created please run the code and you can send request to create and delete helm releases on your kubernetes cluster.


There exist's a sample request in the root level which can be used to test the code.

# Additional Sources
 In order to learn more about kafka you can go through https://developer.confluent.io/get-started/go.