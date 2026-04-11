package vault

// SecretReader is the interface satisfied by any client capable of reading
// a secret at a given path and returning its key/value data.
type SecretReader interface {
	ReadSecret(path string) (map[string]string, error)
}
