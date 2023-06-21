build:
	rm -rf build/
	mkdir build
	go build -o build/etl
	export $PATH=$PATH/:
	echo make sure to add the build/ folder to your local path

run:
	./build/etl --config $(config) --modules $(modules)