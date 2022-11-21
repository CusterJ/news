<!-- Start docker service -->

sudo systemctl restart docker.service

<!-- Run prod -->

docker-compose up

<!-- Stop containers -->

docker-compose down

<!-- Run dev -->

docker-compose -f docker-compose.local.yml up -d

<!-- Then run local backend -->

go run .