GO ?= go
SRC ?= main.go
OUT ?= ensure-access

${OUT}: ${SRC}
	${GO} build -o ${OUT} ${SRC}
