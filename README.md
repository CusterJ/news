<!-- Start docker service -->

sudo systemctl restart docker.service

<!-- Run prod -->

docker-compose up

<!-- Stop containers -->

docker-compose down

<!-- for remove volumes -->

docker-compose down -v 

<!-- to stop network and remove elastic container -->

docker-compose down --remove-orphans

<!-- Run dev -->

docker-compose -f docker-compose.local.yml up -d

<!-- Then run local backend -->

go run .



docker run --name es01 --net elastic -p 9200:9200 -it docker.elastic.co/elasticsearch/elasticsearch:8.5.2

docker run --name es01 --net elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:8.5.2

docker run -d --name elasticsearch -p 9200:9200 -e discovery.type=single-node -v elasticsearch:/usr/share/elasticsearch/data docker.elastic.co/elasticsearch/elasticsearch:8.5.2


docker run \
      --name elasticsearch \
      --net elastic \
      -p 9200:9200 \
      -e discovery.type=single-node \
      -e ES_JAVA_OPTS="-Xms1g -Xmx1g"\
      -e xpack.security.enabled=false \
      -it \
      docker.elastic.co/elasticsearch/elasticsearch:8.5.2