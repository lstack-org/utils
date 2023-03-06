package gt

type gt interface {
	GET()
	POST()
	DELETE()
	PUT()
	UPDATE()
	HEADER()

	InTo()

	AddDecoder()
}

type AddDecoder interface {
	AddJsonDecoder()
	AddYamlDecoder()
}
