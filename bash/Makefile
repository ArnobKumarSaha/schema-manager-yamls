

all: mongo enterprise init-docker

init-docker:
	/home/arnob/files/bash/init-rebuild.sh

enterprise:
	/home/arnob/files/bash/ep-rebuild.sh

mongo:
	/home/arnob/files/bash/mongo-rebuild.sh

dangle:
	y=$(ls)
	echo "$y"
	x=$(docker images -f dangling=true -q)
	docker rmi $x