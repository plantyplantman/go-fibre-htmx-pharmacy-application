package bigc

type MultiStoreResult struct{}

type MultiStoreDoer interface {
	Do() *MultiStoreResult
}

type multiStoreDoer struct {
	client *Client
	data   [][]any
}

func (msd *multiStoreDoer) Do() *MultiStoreResult {
	return &MultiStoreResult{}
}

func NewMultiStoreDoer(client *Client, data [][]any) MultiStoreDoer {
	return &multiStoreDoer{
		client: client,
		data:   data,
	}
}
