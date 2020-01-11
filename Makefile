build:
	GOOS=linux go build
	
run:
	GOOS=linux go build
	RUNEWIDTH_EASTASIAN=1 ./xtelnet