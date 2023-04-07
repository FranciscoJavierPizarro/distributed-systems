module practica2

replace ms => ./ms

replace ra => ./ra

replace escritor => ./cmd/escritor

replace lector => ./cmd/lector

go 1.19

require (
	escritor v0.0.0-00010101000000-000000000000
	lector v0.0.0-00010101000000-000000000000
	ms v0.0.0-00010101000000-000000000000
)

require (
	github.com/DistributedClocks/GoVector v0.0.0-20210402100930-db949c81a0af // indirect
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.1.4 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	ra v0.0.0-00010101000000-000000000000 // indirect
)
