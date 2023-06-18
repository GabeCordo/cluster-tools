build:
	rm -rf build/
	mkdir build
	go build -o build/etl
	export $PATH=$PATH/:
	echo make sure to add the build/ folder to your local path

run:
	./build/etl --config "./.bin/configs/config.etl.yaml" --modules "/Users/gabecordovado/Desktop/EtlTestFiles/modules"

docker:
	docker build . -t gabecordo/etl

docker-run:
	docker run -p 8136:8136 -dit etl