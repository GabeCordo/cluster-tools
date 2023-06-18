local-build:
	rm -rf build/
	mkdir build
	go build -o build/etl
	export $PATH=$PATH/:
	echo make sure to add the build/ folder to your local path

docker-build:
	docker build . -t etl

docker-run:
	docker run -p 8136:8136 -dit etl