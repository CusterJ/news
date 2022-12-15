<!-- Start docker service -->

sudo systemctl restart docker.service

<!-- Run prod -->

docker-compose up --build

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
