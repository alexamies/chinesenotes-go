package transtools

// ApiClient is an interface for translation API clients
type ApiClient interface {
	Translate(sourceText string) (*string, error)
}
